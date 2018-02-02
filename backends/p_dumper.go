package backends

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/flashmob/go-guerrilla/mail"
)

// ----------------------------------------------------------------------------------
// Processor Name: dumper
// ----------------------------------------------------------------------------------
// Description   : Dumps received emails into files
// ----------------------------------------------------------------------------------
func init() {
	processors["dumper"] = func() Decorator {
		return Dumper()
	}
}

type dumperConfig struct {
	DumperDir string `json:"dumper_dir"`
}

func Dumper() Decorator {
	var config *dumperConfig
	initFunc := InitializeWith(func(backendConfig BackendConfig) error {
		configType := BaseConfig(&dumperConfig{})
		bcfg, err := Svc.ExtractConfig(backendConfig, configType)
		if err != nil {
			return err
		}
		config = bcfg.(*dumperConfig)
		return nil
	})
	Svc.AddInitializer(initFunc)
	return func(p Processor) Processor {
		return ProcessWith(func(e *mail.Envelope, task SelectTask) (Result, error) {
			if task == TaskSaveMail {
				uuid, _ := newUUID()
				file := uuid + ".eml"
				err := ioutil.WriteFile(config.DumperDir+"/"+file, e.Data.Bytes(), 0644)

				if err != nil {
					Log().Errorf("Could not dump message to a file - %s", err.Error())
				}

				// continue to the next Processor in the decorator stack
				return p.Process(e, task)
			} else {
				return p.Process(e, task)
			}
		})
	}
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
