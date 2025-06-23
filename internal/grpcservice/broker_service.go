package grpcservice

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hahahamid/broker-backend/config"
	"github.com/hahahamid/broker-backend/internal/repository"
	"github.com/hahahamid/broker-backend/internal/utils"
	pb "github.com/hahahamid/broker-backend/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BrokerService struct {
	pb.UnimplementedBrokerServer
	repo repository.UserRepo
	cfg  *config.Config
}

func NewBrokerService(repo repository.UserRepo, cfg *config.Config) *BrokerService {
	return &BrokerService{repo: repo, cfg: cfg}
}

func (s *BrokerService) Signup(ctx context.Context, req *pb.SignupRequest) (*pb.Empty, error) {
	if err := s.repo.CreateUser(ctx, req.Email, req.Password); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return &pb.Empty{}, nil
}

func (s *BrokerService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	at, rt, err := utils.GenerateTokens(user.ID.Hex(), s.cfg.JWTSecret, s.cfg.RefreshSecret, s.cfg.AccessTokenExpireMin)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "token generation failed")
	}
	_ = s.repo.SaveRefreshToken(ctx, user.ID.Hex(), rt)
	return &pb.AuthResponse{AccessToken: at, RefreshToken: rt}, nil
}

func (s *BrokerService) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.AuthResponse, error) {
	token, err := utils.ValidateToken(req.RefreshToken, s.cfg.RefreshSecret)
	if err != nil || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	at, rt, err := utils.GenerateTokens(userID, s.cfg.JWTSecret, s.cfg.RefreshSecret, s.cfg.AccessTokenExpireMin)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "token generation failed")
	}
	_ = s.repo.SaveRefreshToken(ctx, userID, rt)
	return &pb.AuthResponse{AccessToken: at, RefreshToken: rt}, nil
}

func (s *BrokerService) GetHoldings(ctx context.Context, _ *pb.Empty) (*pb.HoldingsResponse, error) {
	return &pb.HoldingsResponse{
		Holdings: []*pb.Holding{
			{Symbol: "AAPL", Quantity: 10, AvgPrice: 150},
			{Symbol: "GOOGL", Quantity: 5, AvgPrice: 2500},
		},
	}, nil
}

func (s *BrokerService) GetOrderbook(ctx context.Context, _ *pb.Empty) (*pb.OrderbookResponse, error) {
	return &pb.OrderbookResponse{
		Orders: []*pb.Order{
			{Id: "1", Symbol: "AAPL", Side: "buy", Quantity: 10, Price: 150, RealizedPnl: 0, UnrealizedPnl: 20},
			{Id: "2", Symbol: "TSLA", Side: "sell", Quantity: 2, Price: 700, RealizedPnl: 50, UnrealizedPnl: 0},
		},
	}, nil
}

func (s *BrokerService) GetPositions(ctx context.Context, _ *pb.Empty) (*pb.PositionsResponse, error) {
	return &pb.PositionsResponse{
		Positions: []*pb.Position{
			{Symbol: "AAPL", Quantity: 10, AvgPrice: 150, Pnl: 20},
			{Symbol: "TSLA", Quantity: 2, AvgPrice: 700, Pnl: 50},
		},
	}, nil
}
