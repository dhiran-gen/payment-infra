package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoCache struct {
	Client *dynamodb.Client
	Table  string
}

func NewDynamoCache(ctx context.Context, endpoint, table string) (*DynamoCache, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint, SigningRegion: region}, nil
			},
		)),
	)
	if err != nil {
		return nil, err
	}
	client := dynamodb.NewFromConfig(cfg)
	return &DynamoCache{Client: client, Table: table}, nil
}

func (c *DynamoCache) PutPayment(id string, amount int) error {
	_, err := c.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &c.Table,
		Item: map[string]types.AttributeValue{
			"id":     &types.AttributeValueMemberS{Value: id},
			"amount": &types.AttributeValueMemberN{Value: string(rune(amount))},
		},
	})
	return err
}

func (c *DynamoCache) GetPayment(id string) (int, error) {
	out, err := c.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &c.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return 0, err
	}
	if out.Item == nil {
		return 0, nil
	}
	amountAttr, ok := out.Item["amount"].(*types.AttributeValueMemberN)
	if !ok {
		return 0, nil
	}
	// Convert string to int
	var amount int
	_, _ = fmt.Sscanf(amountAttr.Value, "%d", &amount)
	return amount, nil
}
