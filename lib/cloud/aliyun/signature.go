package aliyun

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func WrapRequest(method, baseURI string, reqParams any, apiVersion,
	regionId, accessId, secretKey string, body io.Reader) (*http.Request, error) {
	queryStr := GetQueryString(reqParams, apiVersion, regionId, accessId)
	signStr := method + "&" + URLEncode("/") + "&" + URLEncode(queryStr)
	signKey := secretKey + "&"
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(signKey))
	_, err := io.WriteString(h, signStr)
	if err != nil {
		return nil, err
	}
	tempStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	signedStr := URLEncode(tempStr)
	reqURL := baseURI + "?" + queryStr + "&" + "Signature=" + signedStr
	request, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}
	return request, nil
}

type CommonParam struct {
	Format           string // return JSON or XML，default XML
	Version          string // API Version，format：YYYY-MM-DD， 2017-07-11
	AccessKeyId      string // AccessKeyId
	SignatureMethod  string // SignatureMethod default HMAC-SHA1
	SignatureNonce   string // SignatureNonce
	SignatureVersion string // 1.0
}

func GetStructFieldStr(i interface{}) map[string] /*name*/ string /*value*/ {
	params := make(map[string]string, 5)
	iterStructField(i, params)
	return params
}

func iterStructField(i interface{}, m map[string]string) {
	s := reflect.ValueOf(i)
	t := reflect.TypeOf(i)
	for i := 0; i < s.NumField(); i++ {

		if s.Type().Field(i).Anonymous {
			iterStructField(s.Field(i).Interface(), m)
			continue
		}
		name := s.Type().Field(i).Name
		if s.Field(i).Kind() == reflect.Interface {
			data, _ := json.Marshal(s.Field(i).Interface())
			m[name] = string(data)
		} else if s.Field(i).Kind() == reflect.Map {
			rm, ok := s.Field(i).Interface().(map[string]string)
			if ok {
				for k, v := range rm {
					m[k] = v
				}
			}
		} else {
			v := fmt.Sprintf("%v", s.Field(i).Interface())
			if !s.Field(i).IsValid() {
				continue
			}

			if s.Field(i).Type().Kind() == reflect.Ptr {
				if s.Field(i).IsNil() {
					continue
				}
			}
			tags := t.Field(i).Tag.Get("json")
			if len(tags) > 0 {
				tagnames := strings.Split(tags, ",")
				if len(tagnames) <= 1 {
					name = tags
				} else {
					name = tagnames[0]
				}
			}
			m[name] = v
		}
	}
}

func GetQueryString(reqParams any, apiVersion, _, accessId string) string {
	rp := GetStructFieldStr(reqParams)

	r := rand.Int()
	rpbyte, _ := json.Marshal(rp)
	_hash := md5.New()
	_hash.Write(rpbyte)
	_hash.Write([]byte(strconv.Itoa(r)))
	signonce := uuid.New().String()
	param := CommonParam{
		Format:           "JSON",
		Version:          apiVersion,
		AccessKeyId:      accessId,
		SignatureMethod:  "HMAC-SHA1",
		SignatureVersion: "1.0",
		SignatureNonce:   signonce,
	}
	cp := GetStructFieldStr(param)

	l := len(cp) + len(rp)
	params := make(map[string]string, l)

	for k, v := range cp {
		params[k] = v
	}

	for k, v := range rp {
		params[k] = v
	}

	hs := NewHeaderSorter(params)
	hs.Sort()

	pstrs := make([]string, 0, hs.Len())
	for i := 0; i < hs.Len(); i++ {
		k := URLEncode(hs.Keys[i])
		v := URLEncode(hs.Vals[i])
		pstrs = append(pstrs, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(pstrs, "&")
}

func URLEncode(s string) string {
	s = url.QueryEscape(s)
	return strings.Replace(s, "+", "%20", -1)
}

type HeaderSorter struct {
	Keys []string
	Vals []string
}

// Additional function for function SignHeader.
func NewHeaderSorter(m map[string]string) *HeaderSorter {
	hs := &HeaderSorter{
		Keys: make([]string, 0, len(m)),
		Vals: make([]string, 0, len(m)),
	}

	for k, v := range m {
		hs.Keys = append(hs.Keys, k)
		hs.Vals = append(hs.Vals, v)
	}
	return hs
}

func (hs *HeaderSorter) Sort() { sort.Sort(hs) }

func (hs *HeaderSorter) Len() int {
	return len(hs.Vals)
}

func (hs *HeaderSorter) Less(i, j int) bool {
	return bytes.Compare([]byte(hs.Keys[i]), []byte(hs.Keys[j])) < 0
}

func (hs *HeaderSorter) Swap(i, j int) {
	hs.Vals[i], hs.Vals[j] = hs.Vals[j], hs.Vals[i]
	hs.Keys[i], hs.Keys[j] = hs.Keys[j], hs.Keys[i]
}
