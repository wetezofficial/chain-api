package chainlinktest

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"
)

const base10 = 10

type Head struct {
	ID            uint64
	Hash          common.Hash
	Number        int64
	L1BlockNumber Int64
	ParentHash    common.Hash
	Parent        *Head
	EVMChainID    *Big
	Timestamp     time.Time
	CreatedAt     time.Time
	BaseFeePerGas *Big
}

func (h *Head) UnmarshalJSON(bs []byte) error {
	type head struct {
		Hash          common.Hash    `json:"hash"`
		Number        *hexutil.Big   `json:"number"`
		ParentHash    common.Hash    `json:"parentHash"`
		Timestamp     hexutil.Uint64 `json:"timestamp"`
		L1BlockNumber *hexutil.Big   `json:"l1BlockNumber"`
		BaseFeePerGas *Big           `json:"baseFeePerGas"`
	}

	var jsonHead head
	err := json.Unmarshal(bs, &jsonHead)
	if err != nil {
		return err
	}

	if jsonHead.Number == nil {
		*h = Head{}
		return nil
	}

	h.Hash = jsonHead.Hash
	h.Number = (*big.Int)(jsonHead.Number).Int64()
	h.ParentHash = jsonHead.ParentHash
	h.Timestamp = time.Unix(int64(jsonHead.Timestamp), 0).UTC()
	h.BaseFeePerGas = (*Big)(jsonHead.BaseFeePerGas)
	if jsonHead.L1BlockNumber != nil {
		h.L1BlockNumber = Int64From((*big.Int)(jsonHead.L1BlockNumber).Int64())
	}
	return nil
}

func (h *Head) MarshalJSON() ([]byte, error) {
	type head struct {
		Hash       *common.Hash    `json:"hash,omitempty"`
		Number     *hexutil.Big    `json:"number,omitempty"`
		ParentHash *common.Hash    `json:"parentHash,omitempty"`
		Timestamp  *hexutil.Uint64 `json:"timestamp,omitempty"`
	}

	var jsonHead head
	if h.Hash != (common.Hash{}) {
		jsonHead.Hash = &h.Hash
	}
	jsonHead.Number = (*hexutil.Big)(big.NewInt(int64(h.Number)))
	if h.ParentHash != (common.Hash{}) {
		jsonHead.ParentHash = &h.ParentHash
	}
	if h.Timestamp != (time.Time{}) {
		t := hexutil.Uint64(h.Timestamp.UTC().Unix())
		jsonHead.Timestamp = &t
	}
	return json.Marshal(jsonHead)
}

// Int64 encapsulates the value and validity (not null) of a int64 value,
// to differentiate nil from 0 in json and sql.
type Int64 struct {
	Int64 int64
	Valid bool
}

// NewInt64 returns an instance of Int64 with the passed parameters.
func NewInt64(i int64, valid bool) Int64 {
	return Int64{
		Int64: i,
		Valid: valid,
	}
}

// Int64From creates a new Int64 that will always be valid.
func Int64From(i int64) Int64 {
	return NewInt64(i, true)
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Int.
func (i *Int64) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		// Unmarshal again, directly to value, to avoid intermediate float64
		err = json.Unmarshal(data, &i.Int64)
	case string:
		str := string(x)
		if len(str) == 0 {
			i.Valid = false
			return nil
		}
		i.Int64, err = parse64(str)
	case nil:
		i.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Int64", reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil
	return err
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Int64 if the input is a blank or not an integer.
// It will return an error if the input is not an integer, blank, or "null".
func (i *Int64) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		i.Valid = false
		return nil
	}
	var err error
	i.Int64, err = parse64(string(text))
	i.Valid = err == nil
	return err
}

func parse64(str string) (int64, error) {
	v, err := strconv.ParseInt(str, 10, 64)
	return v, err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Int64 is null.
func (i Int64) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(int64(i.Int64), 10)), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Int64 is null.
func (i Int64) MarshalText() ([]byte, error) {
	if !i.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatInt(int64(i.Int64), 10)), nil
}

// SetValid changes this Int64's value and also sets it to be non-null.
func (i *Int64) SetValid(n int64) {
	i.Int64 = n
	i.Valid = true
}

// Value returns this instance serialized for database storage.
func (i Int64) Value() (driver.Value, error) {
	if !i.Valid {
		return nil, nil
	}

	// golang's sql driver types as determined by IsValue only supports:
	// []byte, bool, float64, int64, string, time.Time
	// https://golang.org/src/database/sql/driver/types.go
	return int64(i.Int64), nil
}

// Scan reads the database value and returns an instance.
func (i *Int64) Scan(value interface{}) error {
	if value == nil {
		*i = Int64{}
		return nil
	}

	switch typed := value.(type) {
	case int:
		safe := int64(typed)
		*i = Int64From(safe)
	case int32:
		safe := int64(typed)
		*i = Int64From(safe)
	case int64:
		safe := int64(typed)
		*i = Int64From(safe)
	case uint:
		if typed > uint(math.MaxInt64) {
			return fmt.Errorf("unable to convert %v of %T to Int64; overflow", value, value)
		}
		safe := int64(typed)
		*i = Int64From(safe)
	case uint64:
		if typed > uint64(math.MaxInt64) {
			return fmt.Errorf("unable to convert %v of %T to Int64; overflow", value, value)
		}
		safe := int64(typed)
		*i = Int64From(safe)
	default:
		return fmt.Errorf("unable to convert %v of %T to Int64", value, value)
	}
	return nil
}

// Big stores large integers and can deserialize a variety of inputs.
type Big big.Int

// NewBig constructs a Big from *big.Int.
func NewBig(i *big.Int) *Big {
	if i != nil {
		var b big.Int
		b.Set(i)
		return (*Big)(&b)
	}
	return nil
}

// NewBigI constructs a Big from int64.
func NewBigI(i int64) *Big {
	return NewBig(big.NewInt(i))
}

// MarshalText marshals this instance to base 10 number as string.
func (b Big) MarshalText() ([]byte, error) {
	return []byte((*big.Int)(&b).Text(base10)), nil
}

// MarshalJSON marshals this instance to base 10 number as string.
func (b Big) MarshalJSON() ([]byte, error) {
	text, err := b.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Big) UnmarshalText(input []byte) error {
	input = RemoveQuotes(input)
	str := string(input)
	if HasHexPrefix(str) {
		decoded, err := hexutil.DecodeBig(str)
		if err != nil {
			return err
		}
		*b = Big(*decoded)
		return nil
	}

	_, ok := b.setString(str, 10)
	if !ok {
		return fmt.Errorf("unable to convert %s to Big", str)
	}
	return nil
}

func (b *Big) setString(s string, base int) (*Big, bool) {
	w, ok := (*big.Int)(b).SetString(s, base)
	return (*Big)(w), ok
}

// UnmarshalJSON implements encoding.JSONUnmarshaler.
func (b *Big) UnmarshalJSON(input []byte) error {
	return b.UnmarshalText(input)
}

// Value returns this instance serialized for database storage.
func (b Big) Value() (driver.Value, error) {
	return b.String(), nil
}

// Scan reads the database value and returns an instance.
func (b *Big) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		decoded, ok := b.setString(v, 10)
		if !ok {
			return fmt.Errorf("unable to set string %v of %T to base 10 big.Int for Big", value, value)
		}
		*b = *decoded
	case []uint8:
		// The SQL library returns numeric() types as []uint8 of the string representation
		decoded, ok := b.setString(string(v), 10)
		if !ok {
			return fmt.Errorf("unable to set string %v of %T to base 10 big.Int for Big", value, value)
		}
		*b = *decoded
	default:
		return fmt.Errorf("unable to convert %v of %T to Big", value, value)
	}

	return nil
}

// ToInt converts b to a big.Int.
func (b *Big) ToInt() *big.Int {
	return (*big.Int)(b)
}

// String returns the base 10 encoding of b.
func (b *Big) String() string {
	return b.ToInt().String()
}

// Hex returns the hex encoding of b.
func (b *Big) Hex() string {
	return hexutil.EncodeBig(b.ToInt())
}

// Cmp compares b and c as big.Ints.
func (b *Big) Cmp(c *Big) int {
	return b.ToInt().Cmp(c.ToInt())
}

// Equal returns true if c is equal according to Cmp.
func (b *Big) Equal(c *Big) bool {
	return b.Cmp(c) == 0
}

// IsQuoted checks if the first and last characters are either " or '.
func IsQuoted(input []byte) bool {
	return len(input) >= 2 &&
		((input[0] == '"' && input[len(input)-1] == '"') ||
			(input[0] == '\'' && input[len(input)-1] == '\''))
}

// RemoveQuotes removes the first and last character if they are both either
// " or ', otherwise it is a noop.
func RemoveQuotes(input []byte) []byte {
	if IsQuoted(input) {
		return input[1 : len(input)-1]
	}
	return input
}

// HasHexPrefix returns true if the string starts with 0x.
func HasHexPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}
