package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/ethpandaops/beacon/pkg/beacon"
	"github.com/golang/snappy"
	"github.com/sirupsen/logrus"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	addr := getEnv("BEACON_ADDR", "localhost")
	port := getEnv("BEACON_PORT", "5052")
	slot := os.Args[1]

	logger := logrus.New()

	raw(logger, addr, port, slot)
	eth2Client(logger, addr, port, slot)
}

func raw(logger *logrus.Logger, addr, port, slot string) {
	nestedLogger := logger.WithField("source", "raw")
	url := fmt.Sprintf("http://%s:%s/eth/v1/beacon/blocks/%s", addr, port, slot)

	resp, err := http.Get(url)
	if err != nil {
		nestedLogger.Errorf("Failed to get block data: %v", err)
		return
	}
	defer resp.Body.Close()

	blockData, err := io.ReadAll(resp.Body)
	if err != nil {
		nestedLogger.Errorf("Failed to read block data: %v", err)
		return
	}

	var blockDataMap map[string]interface{}
	if err := json.Unmarshal(blockData, &blockDataMap); err != nil {
		nestedLogger.Errorf("Failed to unmarshal block data: %v", err)
		return
	}

	data, ok := blockDataMap["data"].(map[string]interface{})
	if !ok {
		nestedLogger.Errorf("Failed to get data from block data")
		return
	}

	messageData, ok := data["message"]
	if !ok {
		nestedLogger.Errorf("Failed to get message data from block data")
		return
	}

	messageDataAsJSON, err := json.Marshal(messageData)
	if err != nil {
		nestedLogger.Errorf("Failed to marshal message data: %v", err)
		return
	}

	logDataSizes(nestedLogger, messageDataAsJSON)
}

func eth2Client(logger *logrus.Logger, addr, port, slot string) {
	nestedLogger := logger.WithField("source", "eth2Client")
	opts := *beacon.DefaultOptions()
	emptyLogger := logrus.New()
	emptyLogger.SetLevel(logrus.PanicLevel)
	beaconNode := beacon.NewNode(emptyLogger, &beacon.Config{
		Addr: fmt.Sprintf("%s:%s", addr, port),
		Name: "beacon node",
	}, "eth", opts)

	if err := beaconNode.Start(context.Background()); err != nil {
		nestedLogger.Errorf("Failed to start beacon node: %v", err)
		return
	}

	block, err := beaconNode.FetchBlock(context.Background(), slot)
	if err != nil {
		nestedLogger.Errorf("Failed to fetch block: %v", err)
		return
	}

	blockMessage, err := getBlockMessage(block)
	if err != nil {
		nestedLogger.Errorf(err.Error())
		return
	}

	dataAsJSON, err := json.Marshal(blockMessage)
	if err != nil {
		nestedLogger.Errorf("Failed to marshal block data: %v", err)
		return
	}

	logDataSizes(nestedLogger, dataAsJSON)
}

func getBlockMessage(block *spec.VersionedSignedBeaconBlock) (interface{}, error) {
	switch block.Version {
	case spec.DataVersionAltair:
		return block.Altair.Message, nil
	case spec.DataVersionBellatrix:
		return block.Bellatrix.Message, nil
	case spec.DataVersionCapella:
		return block.Capella.Message, nil
	case spec.DataVersionDeneb:
		return block.Deneb.Message, nil
	default:
		return nil, fmt.Errorf("Unsupported block version: %s", block.Version)
	}
}

func logDataSizes(logger *logrus.Entry, data []byte) {
	dataSize := len(data)
	compressedData := snappy.Encode(nil, data)
	compressedDataSize := len(compressedData)
	logger.Infof("Byte length of JSON block data: %d", dataSize)
	logger.Infof("Compressed byte length of JSON block data: %d", compressedDataSize)
}
