package serial

import (
	"encoding/csv"
	"os"	
	"fmt"
	"strconv"
	"io"
)

type BetReader struct {
	batchSize int32
	file *os.File
	csvReader *csv.Reader
	currentBetCount int32
}

func NewBetReader(batchSize int32, file string) (*BetReader, error) {

	betsFile, err := os.Open(file)

	if err != nil {
		return nil, fmt.Errorf("Error opening bets file: %w", err)
	}
	


	reader := &BetReader{
		batchSize: batchSize,
		file: betsFile,
		csvReader: csv.NewReader(betsFile),
	}
	return reader, nil
}

func (reader *BetReader) yieldBet(packetBuilder *PacketBuilder) (bool, error) {
	record, err := reader.csvReader.Read()
	if err != nil {
		return false, err
	}

	// Asummed format name, surname , dni, birth, number

	if len(record) < 5 {
		return false, fmt.Errorf("invalid record: expected 5 fields, got %d", len(record))
	}

	dni, err := strconv.ParseInt(record[2], 10, 32)
	if err != nil {
		return false, fmt.Errorf("invalid DNI: %w", err)
	}

	number, err := strconv.ParseInt(record[4], 10, 32)
	if err != nil {
		return false, fmt.Errorf("invalid Number: %w", err)
	}

	bet := PersonBet{
		Name:    record[0],
		Surname: record[1],
		Dni:     int32(dni),
		Birth:   record[3],
		Number:  int32(number),
	}

	ok := packetBuilder.WriteBet(&bet)
	return ok, nil
}

func (reader *BetReader) YieldBatch(packetBuilder *PacketBuilder) (int32, error) {

	// This doesnt work properly when faced with record of size 8kb+ and batchSize == 1
	continueReading, err := reader.yieldBet(packetBuilder)
	reader.currentBetCount +=1 

	for err == nil && continueReading && reader.currentBetCount < reader.batchSize {
		continueReading, err = reader.yieldBet(packetBuilder)
		reader.currentBetCount +=1 
	}

	readedCount:= reader.currentBetCount
	if continueReading{
		reader.currentBetCount = 0
	} else if err == nil {
		readedCount-=1 // Subtract the last bet since it was left for the next batch.
		reader.currentBetCount = 1 // Remainder bet counted. This could be on the PacketBuilder but better to separate concerns.
	} else if err == io.EOF {
		// Ignore EOF error, the next yield will throw readCount == 0 allegedly. But send last readedCount bets. 
		err = nil
		reader.currentBetCount = 0		
		readedCount-=1 // Subtract the last bet since the EOF was counted as a bet.
	}

	return readedCount, err
}

func (reader *BetReader) Close() error {
	return reader.file.Close()
}