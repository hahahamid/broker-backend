package main

import (
	"context"
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

		// If you omit version, it defaults to IPv4.
		// If you want IPv6, use ?version=ipv6

		version := c.DefaultQuery("version", "ipv4")

		var (
			url    string
			source string
		)
		switch version {
		case "ipv6":
			url = "https://api64.ipify.org?format=json"
			source = "ipv6"
		default:
			url = "https://api.ipify.org?format=json"
			source = "ipv4"
		}

		resp, err := ipClient.Get(url)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		var body struct {
			IP string `json:"ip"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("welcome to the server from %s", body.IP),
			"source":  source,
		})
	})

	log.Println("HTTP server @ :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Gin run: %v", err)
	}
}
