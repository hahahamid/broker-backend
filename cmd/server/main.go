package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hahahamid/broker-backend/config"
	grpcService "github.com/hahahamid/broker-backend/internal/grpcservice"
	"github.com/hahahamid/broker-backend/internal/handlers"
	"github.com/hahahamid/broker-backend/internal/middleware"
	"github.com/hahahamid/broker-backend/internal/repository"
	pb "github.com/hahahamid/broker-backend/proto"
)

var ipClient = &http.Client{
	Timeout: 3 * time.Second,
}

type ipResponse struct {
	IP string `json:"ip"`
}

// this func will first try the IPv6 endpoint, and if it errors (timeout,
// no connectivity, etc), it will go to the IPv4 endpoint.

func getPublicIP(ctx context.Context) (ip, source string, err error) {
	endpoints := []struct {
		url, src string
	}{
		{"https://api64.ipify.org?format=json", "ipv6"},
		{"https://api.ipify.org?format=json", "ipv4"},
	}

	for _, ep := range endpoints {
		reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		req, _ := http.NewRequestWithContext(reqCtx, http.MethodGet, ep.url, nil)
		resp, err := ipClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var body ipResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			continue
		}
		if body.IP == "" {
			continue
		}
		return body.IP, ep.src, nil
	}

	return "", "", errors.New("could not get public IP (both ipv6 & ipv4 failed)")
}

func main() {
	cfg := config.Load()
	repo, err := repository.NewMongoRepo(cfg)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}

	// 1️⃣ Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("gRPC listen: %v", err)
		}
		grpcServer := grpcLib.NewServer()
		pb.RegisterBrokerServer(grpcServer, grpcService.NewBrokerService(repo, cfg))
		log.Println("gRPC server @ :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve: %v", err)
		}
	}()

	// 2️⃣ Start HTTP→gRPC gateway
	go func() {
		ctx := context.Background()
		mux := runtime.NewServeMux()
		opts := []grpcLib.DialOption{grpcLib.WithTransportCredentials(insecure.NewCredentials())}
		if err := pb.RegisterBrokerHandlerFromEndpoint(ctx, mux, "localhost:50051", opts); err != nil {
			log.Fatalf("gateway register: %v", err)
		}
		log.Println("gRPC-Gateway @ :8081")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Fatalf("gateway serve: %v", err)
		}
	}()

	// 3️⃣ Existing HTTP+Gin server
	r := gin.Default()
	ah := handlers.NewAuthHandler(repo, cfg)
	hh := handlers.NewHoldingsHandler()
	ob := handlers.NewOrderbookHandler()
	ph := handlers.NewPositionsHandler()

	r.GET("/health", func(c *gin.Context) { c.Status(200) })
	r.POST("/signup", ah.Signup)
	r.POST("/login", ah.Login)
	r.POST("/refresh", ah.Refresh)

	auth := r.Group("/", middleware.JWTAuth(cfg))
	{
		auth.GET("/holdings", hh.Get)
		auth.GET("/orderbook", ob.Get)
		auth.GET("/positions", ph.Get)
	}

	// POST API AS REQUESTED

	r.POST("/push-data", func(c *gin.Context) {
		ip, source, err := getPublicIP(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("welcome to the server from %s", ip),
			"source":  source,
		})
	})

	log.Println("HTTP server @ :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Gin run: %v", err)
	}
}
