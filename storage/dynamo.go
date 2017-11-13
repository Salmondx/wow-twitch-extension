package storage

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/salmondx/wow-twitch-extension/model"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const characterTable = "STREAMER_CHARACTERS"
const characterLimit = 20

type CharacterInfoItem struct {
	*model.CharacterInfo
	CharacterID string `json:"characterID"`
	StreamerID  string `json:"streamerID"`
}

// DynamoRepository is a CharacterRepository implementation for DynamoDB
type DynamoRepository struct {
	client *dynamodb.DynamoDB
}

func New() (*DynamoRepository, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		return nil, fmt.Errorf("Can not create dynamodb client: %v", err)
	}
	client := dynamodb.New(sess)
	return &DynamoRepository{client}, nil
}

func (db *DynamoRepository) List(streamerID string) ([]*model.CharacterInfo, error) {
	if streamerID == "" {
		return nil, errors.New("streamerID can not be empty")
	}

	query := selectAllQuery(streamerID)
	resp, err := db.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Can not get characters for %s, reason: %v", streamerID, err)
	}

	characterInfos := make([]*model.CharacterInfo, len(resp.Items))

	for i, item := range resp.Items {
		characterItem := &CharacterInfoItem{}

		err = dynamodbattribute.UnmarshalMap(item, characterItem)
		if err != nil {
			return nil, fmt.Errorf("Can not unmarshal result: %v", err)
		}

		characterInfos[i] = characterItem.CharacterInfo
	}
	return characterInfos, nil
}

func (db *DynamoRepository) Add(streamerID string, character *model.CharacterInfo) error {
	if streamerID == "" || character == nil {
		return errors.New("StreamerID or character info can not be empty")
	}

	query := selectAllQuery(streamerID)
	query.SetSelect("COUNT")

	resp, err := db.client.Query(query)
	if err != nil {
		return fmt.Errorf("Can not count characters for %s. Reason: %v", streamerID, err)
	}
	count := resp.Count
	if *count >= characterLimit {
		return model.CharacterLimitError{fmt.Sprintf("Can't add character for %s. Limit is 20.", streamerID)}
	}

	characterItem := CharacterInfoItem{
		CharacterInfo: character,
		CharacterID:   createCharacterID(character),
		StreamerID:    streamerID,
	}

	req, err := dynamodbattribute.MarshalMap(characterItem)
	if err != nil {
		return fmt.Errorf("Can not marshal character item: %v. Reason: %v", characterItem, err)
	}

	_, err = db.client.PutItem(&dynamodb.PutItemInput{
		Item:      req,
		TableName: aws.String(characterTable),
	})
	if err != nil {
		return fmt.Errorf("Can not insert item into db. Reason: %v", err)
	}

	return nil
}

func (db *DynamoRepository) Delete(streamerID, realm, name string) error {
	if streamerID == "" || realm == "" || name == "" {
		return errors.New("StreamerID, realm or name can not be empty")
	}

	query := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"streamerID": {
				S: aws.String(streamerID),
			},
			"characterID": {
				S: aws.String(genCharacterID(realm, name)),
			},
		},
		TableName: aws.String(characterTable),
	}

	_, err := db.client.DeleteItem(query)
	if err != nil {
		return fmt.Errorf("Can't delete character %s on realm %s of streamer %s. Reason: %v", name, realm, streamerID, err)
	}
	return nil
}

func createCharacterID(character *model.CharacterInfo) string {
	return genCharacterID(character.Realm, character.Name)
}

func genCharacterID(realm, name string) string {
	return realm + ":" + name
}

func selectAllQuery(streamerID string) *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName: aws.String(characterTable),
		KeyConditions: map[string]*dynamodb.Condition{
			"streamerID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: &streamerID,
					},
				},
			},
		},
	}
}
