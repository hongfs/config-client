package config

import (
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/random"
	"github.com/duke-git/lancet/v2/slice"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Load     bool      `json:"load"`
	Value    string    `json:"value"`
	Error    error     `json:"error"`
	LoadTime time.Time `json:"load_time"`
}

func (c *Config) String() string {
	return c.Value
}

func (c *Config) Bytes() []byte {
	return []byte(c.Value)
}

func (c *Config) Int() int {
	return int(c.Int64())
}

func (c *Config) Int64() int64 {
	value, _ := strconv.ParseInt(c.Value, 10, 64)

	return value
}

func (c *Config) Uint64() uint64 {
	return uint64(c.Int64())
}

func (c *Config) Bool() bool {
	return slice.Contain(
		[]string{"1", "true", "yes"},
		c.Value,
	)
}

func (c *Config) Array() []string {
	if strings.Contains(c.Value, ",") {
		return strings.Split(c.Value, ",")
	}

	return []string{c.Value}
}

func (c *Config) Random() string {
	arr := c.Array()

	if len(arr) == 0 {
		return ""
	}

	if len(arr) == 1 {
		return arr[0]
	}

	return arr[random.RandInt(0, len(arr))]
}

func (c *Config) Shuffle() []string {
	arr := c.Array()

	slice.Shuffle(arr)

	return arr
}

func (c *Config) Allow(value string) bool {
	arr := c.Array()

	for _, v := range arr {
		if v == value {
			return true
		}
	}

	return false
}

var cacheConfig = map[string]*Config{}

func Get(name string) *Config {
	value, ok := cacheConfig[name]

	if ok && value != nil && value.LoadTime.Add(time.Minute*1).After(time.Now()) {
		return value
	}

	var config Config

	v, err := Load(name)

	if err != nil {
		config.Error = err
		return &config
	}

	config.Value = v
	config.Load = true
	config.LoadTime = time.Now()

	cacheConfig[name] = &config

	return &config
}

func Load(name string) (string, error) {
	servicePrefix := os.Getenv("CONFIG_SERVICE_PREFIX")

	if servicePrefix == "" {
		return "", errors.New("CONFIG_SERVICE_PREFIX is not set")
	}

	serviceUrl := fmt.Sprintf("%s%s", servicePrefix, name)

	resp, err := http.Get(serviceUrl)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}
