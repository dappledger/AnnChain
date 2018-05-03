package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Bytes []byte

func (b *Bytes) MarshalJSON() ([]byte, error) {
	bys := strings.ToUpper(hex.EncodeToString(*b))
	return json.Marshal(bys)
}

func (b *Bytes) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	var bys []byte
	bys, err = hex.DecodeString(str)
	if err != nil {
		return err
	}
	(*b) = Bytes(bys)
	return nil
}

func (b *Bytes) Bytes() []byte {
	return []byte(*b)
}

func (b *Bytes) String() string {
	ret, err := b.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("marshal err:%v", err)
	}
	return string(ret)
}

/////////////////////////////////////////////////////////////////

const (
	timeFormart = "2006-01-02 15:04:05"
)

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(&t.Time)
}

func (t *Time) UnmarshalJSON(data []byte) error {
	st := struct {
		time.Time
	}{}
	err := json.Unmarshal(data, &st)
	if err != nil {
		return err
	}
	t.Time = st.Time
	return nil
}

func (t *Time) String() string {
	return t.Format(timeFormart)
}
