package gorq

import (
	"bytes"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/goibibo/mantle"
	"github.com/goibibo/t-coredb"
	. "github.com/kisielk/og-rek"
	"github.com/nu7hatch/gouuid"
	"sync"
)

var rqRedisPool mantle.Mantle
var rwMutex sync.RWMutex

type Hargs map[string]string

type RQJob struct {
	JobId    string
	funcName string
	args     []string
	kwargs   Hargs
}

func NewUUID() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return uuid.String()
}

func InitRedisPool(whichRedis string) {
	//protect two guys trying to read rqRedisPool at once
	rwMutex.Lock()
	defer rwMutex.Unlock()
	if rqRedisPool == nil {
		rqRedisPool = db.GetRedisClientFor(whichRedis)
	}
}

func DecodeResult(result string) string {
	//decoding encoded value
	buf := bytes.NewBufferString(result)
	dec := NewDecoder(buf)
	v, err := dec.Decode()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("decoded value", v)
	return v.(string)
}

func (job *RQJob) EncodeJob() string {
	//encoding stuff
	p := &bytes.Buffer{}
	e := NewEncoder(p)
	f := []interface{}{job.funcName, nil, job.args, job.kwargs}
	e.Encode(f)
	fmt.Println("encoded value", string(p.Bytes()))
	return string(p.Bytes())
}

func (job *RQJob) QueueId() string {
	return fmt.Sprintf("rq:job:%s", job.JobId)
}

func (job *RQJob) GetJobId() string {
	return job.JobId
}

func (job *RQJob) EnqueueJob(rqJob Hargs) {
	queueId := job.QueueId()
	_, err := rqRedisPool.Execute("HMSET", redis.Args{queueId}.AddFlat(rqJob)...)
	if err != nil {
		fmt.Println("HMSET", err)
	}
}

func NewRQJob(funcName string, args []string, kwargs Hargs) *RQJob {
	return &RQJob{JobId: NewUUID(), funcName: funcName, args: args, kwargs: kwargs}
}

func (job *RQJob) Enqueue() {
	rqJob := map[string]string{"data": job.EncodeJob()}
	job.EnqueueJob(rqJob)
}

func (job *RQJob) StartJob() {
	_, err := rqRedisPool.Execute("RPUSH", "rq:queue:default", job.JobId)
	if err != nil {
		fmt.Println("RPUSH", err)
	}
}

func (job *RQJob) Result() {
	queueId := job.QueueId()
	fmt.Println("queueId", queueId)
	result, err := redis.String(rqRedisPool.Execute("HGET", queueId, "result"))
	if err != nil {
		fmt.Println("HGETALL", err)
	}
	fmt.Println("values", DecodeResult(result))
}
