package modules

import (
	"fmt"
	"github.com/davidscholberg/go-i3barjson"
	"reflect"
	"time"
)

// BlockConfig is an interface for Block configuration structs.
type BlockConfig interface {
	GetBlockIndex() int
	GetUpdateFunc() func(b *i3barjson.Block, c BlockConfig)
	GetUpdateInterval() float64
	GetUpdateSignal() int
}

// GlobalConfig represents global config options.
type GlobalConfig struct {
	Debug           bool    `yaml:"debug"`
	RefreshInterval float64 `yaml:"refresh_interval"`
}

// GoBlock contains all functions and objects necessary to configure and update
// a block.
type GoBlock struct {
	Block  i3barjson.Block
	Config BlockConfig
	Ticker *time.Ticker
	Update func(b *i3barjson.Block, c BlockConfig)
}

// Config is the root configuration struct.
type Config struct {
	Global GlobalConfig `yaml:"global"`
	Blocks BlockConfigs `yaml:"blocks"`
}

// BlockConfigs holds the configuration of all blocks.
type BlockConfigs struct {
	Disk         Disk          `yaml:"disk"`
	Interfaces   []Interface   `yaml:"interfaces"`
	Load         Load          `yaml:"load"`
	Memory       Memory        `yaml:"memory"`
	Raid         Raid          `yaml:"raid"`
	Temperatures []Temperature `yaml:"temperatures"`
	Time         Time          `yaml:"time"`
	Volume       Volume        `yaml:"volume"`
}

// GetGoBlocks initializes and returns a GoBlock slice based on the
// given configuration.
func GetGoBlocks(c BlockConfigs) ([]*GoBlock, error) {
	// TODO: error handling
	// TODO: include i3barjson.Block config in config structs
	var blockConfigSlice []BlockConfig
	cType := reflect.ValueOf(c)
	for i := 0; i < cType.NumField(); i++ {
		field := cType.Field(i)
		switch field.Kind() {
		case reflect.Struct:
			b := field.Interface().(BlockConfig)
			if b.GetBlockIndex() > 0 {
				blockConfigSlice = append(blockConfigSlice, b)
			}
		case reflect.Slice:
			for i := 0; i < field.Len(); i++ {
				b := field.Index(i).Interface().(BlockConfig)
				if b.GetBlockIndex() > 0 {
					blockConfigSlice = append(blockConfigSlice, b)
				}
			}
		default:
			return nil, fmt.Errorf("unexpected type: %s\n", field.Type())
		}
	}

	goblocks := make([]*GoBlock, len(blockConfigSlice))
	for _, blockConfig := range blockConfigSlice {
		blockIndex := blockConfig.GetBlockIndex()
		updateFunc := blockConfig.GetUpdateFunc()
		ticker := time.NewTicker(
			time.Duration(
				blockConfig.GetUpdateInterval() * float64(time.Second),
			),
		)
		goblocks[blockIndex-1] = &GoBlock{
			i3barjson.Block{Separator: true, SeparatorBlockWidth: 20},
			blockConfig,
			ticker,
			updateFunc,
		}
	}

	return goblocks, nil
}

// SelectAction is a function type that specifies an action to perform when a
// channel is selected on in the main program loop. The first returned bool
// indicates whether or not Goblocks should refresh the output. The second
// returned bool indicates whether or not Goblocks should exit the loop.
type SelectAction func(*GoBlock) (bool, bool)

// SelectCases represents the set of channels that Goblocks selects on in the
// main program loop, as well as the functions and data to run and operate on,
// respectively.
type SelectCases struct {
	Cases   []reflect.SelectCase
	Actions []SelectAction
	Blocks  []*GoBlock
}

// Add adds a channel, action, and GoBlock to the SelectCases object.
func (s *SelectCases) Add(c interface{}, a SelectAction, b *GoBlock) {
	selectCase := reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(c),
	}
	s.Cases = append(s.Cases, selectCase)
	s.Actions = append(s.Actions, a)
	s.Blocks = append(s.Blocks, b)
}

// AddBlockSelectCases is a helper function to add all configured GoBlock
// objects to SelectCases.
func (s *SelectCases) AddBlockSelectCases(b []*GoBlock) {
	for _, goblock := range b {
		addBlockToSelectCase(s, goblock)
	}
}

// AddChanSelectCase is a helper function that adds a non-GoBlock channel and
// action to SelectCases. This can be used for signal handling and other non-
// block specific operations.
func (s *SelectCases) AddChanSelectCase(c interface{}, a SelectAction) {
	s.Add(
		c,
		a,
		nil,
	)
}

// addBlockToSelectCase is a helper function to add a GoBlock to SelectCases.
// The channel used is a time.Ticker channel set to tick according to the
// block's configuration. The SelectAction function updates the block's status
// but does not tell Goblocks to refresh.
func addBlockToSelectCase(s *SelectCases, b *GoBlock) {
	updateFunc := b.Update
	s.Add(
		b.Ticker.C,
		func(b *GoBlock) (bool, bool) {
			updateFunc(&b.Block, b.Config)
			return false, false
		},
		b,
	)
}

// SelectActionExit is a helper function of type SelectAction that tells
// Goblocks to exit.
func SelectActionExit(b *GoBlock) (bool, bool) {
	return false, true
}

// SelectActionRefresh is a helper function of type SelectAction that tells
// Goblocks to refresh the output.
func SelectActionRefresh(b *GoBlock) (bool, bool) {
	return true, false
}
