package grpc

import (
	"context"
	"fmt"
	"migration-service/internal/domain"
	"time"

	structurepb "migration-service/api/proto/structure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// StructureClient implements StructureService using gRPC
type StructureClient struct {
	client structurepb.StructureServiceClient
	conn   *grpc.ClientConn
}

// NewStructureClient creates a new gRPC client for Structure Service
func NewStructureClient(address string) (*StructureClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to structure service: %w", err)
	}

	client := structurepb.NewStructureServiceClient(conn)

	return &StructureClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *StructureClient) Close() error {
	return c.conn.Close()
}

// CreateStructure creates or updates the full structure hierarchy
func (c *StructureClient) CreateStructure(ctx context.Context, data *domain.StructureData) (*domain.StructureResult, error) {
	req := &structurepb.CreateStructureRequest{
		Inn:         data.INN,
		Kpp:         data.KPP,
		Foiv:        data.FOIV,
		OrgName:     data.OrgName,
		BranchName:  data.BranchName,
		FacultyName: data.FacultyName,
		Course:      int32(data.Course),
		GroupNumber: data.GroupNumber,
	}

	resp, err := c.client.CreateStructure(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create structure: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("structure service error: %s", resp.Error)
	}

	result := &domain.StructureResult{
		UniversityID: int(resp.UniversityId),
		GroupID:      int(resp.GroupId),
	}

	if resp.BranchId != nil {
		branchID := int(*resp.BranchId)
		result.BranchID = &branchID
	}

	if resp.FacultyId != nil {
		facultyID := int(*resp.FacultyId)
		result.FacultyID = &facultyID
	}

	return result, nil
}

// LinkGroupToChat links a group to a chat
func (c *StructureClient) LinkGroupToChat(ctx context.Context, groupID int, chatID int) error {
	req := &structurepb.LinkGroupToChatRequest{
		GroupId: int64(groupID),
		ChatId:  int64(chatID),
	}

	resp, err := c.client.LinkGroupToChat(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to link group to chat: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to link group to chat: %s", resp.Error)
	}

	return nil
}
