package model

import (
	"context"
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
	return dao.rdb.HGetAll(ctx, spreadsheetId).Result()
}
