package model

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Dao struct {
	rdb *redis.Client
}

func NewDao(rdb *redis.Client) *Dao {
	return &Dao{
		rdb: rdb,
	}
}

var ctx = context.Background()

func (dao *Dao) IsSpreadsheetExists(spreadsheetId string) (bool, error) {
	val, err := dao.rdb.Exists(ctx, strings.ToLower(spreadsheetId)).Result()
	if err != nil {
		return false, err
	}

	return val == 1, nil
}

func (dao *Dao) GetSpreadeetKeys(spreadsheetId string) ([]string, error) {
	return dao.rdb.HKeys(ctx, strings.ToLower(spreadsheetId)).Result()
}

func (dao *Dao) SetCell(spreadsheetId string, cellId string, value string) error {
	if err := dao.rdb.HSet(ctx, strings.ToLower(spreadsheetId), strings.ToLower(cellId), value).Err(); err != nil {
		return err
	}

	return nil
}

func (dao *Dao) GetCell(spreadsheetId string, cellId string) (string, error) {
	return dao.rdb.HGet(ctx, strings.ToLower(spreadsheetId), strings.ToLower(cellId)).Result()
}

func (dao *Dao) GetAllCells(spreadsheetId string) (map[string]string, error) {
	return dao.rdb.HGetAll(ctx, strings.ToLower(spreadsheetId)).Result()
}

func (dao *Dao) GetDependants(spreadsheetId string, cellId string) ([]string, error) {
	return dao.rdb.SMembers(ctx, strings.ToLower(spreadsheetId)+"/"+strings.ToLower(cellId)).Result()
}

func (dao *Dao) AddDependatFormula(spreadsheetId string, cellId string, dependsOn []string) error {
	for _, dependantCellId := range dependsOn {
		if dependantCellId == cellId {
			continue
		}

		err := dao.rdb.SAdd(ctx, strings.ToLower(spreadsheetId)+"/"+strings.ToLower(dependantCellId), cellId).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (dao *Dao) DeleteDependatFormula(spreadsheetId string, cellId string, dependsOn []string) error {
	for _, dependantCellId := range dependsOn {
		err := dao.rdb.SRem(ctx, strings.ToLower(spreadsheetId)+"/"+strings.ToLower(dependantCellId), cellId).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func subscriptionKey(id string) string {
	return fmt.Sprintf("subscription:%s", id)
}

func (dao *Dao) CreateSubscription(spreadsheetId string, cellId string) (string, error) {
	idVal, err := dao.rdb.Incr(ctx, "subscription:counter").Result()
	if err != nil {
		return "", err
	}

	fmt.Print(idVal)
	id := strconv.FormatInt(idVal, 16)

	if err := dao.rdb.HSet(ctx, subscriptionKey(id),
		"spreadsheetId",
		strings.ToLower(spreadsheetId),
		"cellId",
		strings.ToLower(cellId)).Err(); err != nil {
		return "", err
	}

	return id, nil
}

func subscriptionPubSubKey(spreadsheetId, cellId string) string {
	return fmt.Sprintf("pubsub:%s/%s", spreadsheetId, cellId)
}

var ERROR_NO_SUBSCRIPTION = errors.New("Unknown key")

func (dao *Dao) GetSubscription(subId string) (map[string]string, error) {
	data, err := dao.rdb.HGetAll(ctx, subscriptionKey(subId)).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ERROR_NO_SUBSCRIPTION
	}

	return data, nil
}

func (dao *Dao) Subscribe(subId string) (*redis.PubSub, error) {
	data, err := dao.GetSubscription(subId)
	if err != nil {
		return nil, err
	}

	return dao.rdb.Subscribe(
			ctx,
			subscriptionPubSubKey(data["spreadsheetId"], data["cellId"]),
		),
		nil
}

func (dao *Dao) NotifyCellChange(spreadsheetId, cellId string) error {
	return dao.rdb.Publish(ctx, subscriptionPubSubKey(spreadsheetId, cellId), nil).Err()
}
