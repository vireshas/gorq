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

//A pool of redis connections which enqueues job in redis
var rqRedisPool mantle.Mantle

//Mutex to protect concurrent writes to rqRedisPool
var rwMutex sync.RWMutex

//Syntactic sugar for params that are passed to func
type Hargs map[string]string

/*
 * We will be enqueuing a python job of type
   def funcName(*args, **kwargs)
   args: can be of type string only
   kwargs: this is of type map[string]string
*/
//TODO: enqueues all the jobs in default queue
//      add support to add a job to other queues
type RQJob struct {
	Id       string
	funcName string
	args     []string
	kwargs   Hargs
}

//Creates a new job to enqueue
func NewRQJob(funcName string, args []string, kwargs Hargs) *RQJob {
	return &RQJob{Id: NewUUID(), funcName: funcName, args: args, kwargs: kwargs}
}

//Generates unique id for a job
//TODO: check if we need a distrubted uuid generator so that
//      jobs enqueued from multiple servers are unique
//      Jobs in RQ expire after 442ms to I am assuming this wouldn't be an issue
func NewUUID() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return uuid.String()
}

//This creates a redis pool
func InitRedisPool(vertical string) {
	//protect two guys trying to read rqRedisPool at once
	rwMutex.Lock()
	defer rwMutex.Unlock()
	if rqRedisPool == nil {
		rqRedisPool = db.GetRedisClientFor(vertical)
	}
}

//Results in RQ are pickle encoded; This method decodes the result
func DecodeResult(result string) string {
	//decoding encoded value
	buf := bytes.NewBufferString(result)
	dec := NewDecoder(buf)
	v, err := dec.Decode()
	if err != nil {
		fmt.Println(err)
	}
	return v.(string)
}

//This method encodes a string in pickle format
//RQ expects python func that we are calling to be pickle encoded
func (job *RQJob) EncodeJob() string {
	//encoding stuff
	p := &bytes.Buffer{}
	e := NewEncoder(p)
	f := []interface{}{job.funcName, nil, job.args, job.kwargs}
	e.Encode(f)
	return string(p.Bytes())
}

//This was supposed to be a getter for job's id
//I then made job Id public
//TODO: make id private and add getters, setters
func (job *RQJob) QueueId() string {
	return fmt.Sprintf("rq:job:%s", job.Id)
}

//Set job related metadata to redis
func (job *RQJob) EnqueueJob(rqJob Hargs) {
	queueId := job.QueueId()
	_, err := rqRedisPool.Execute("HMSET", redis.Args{queueId}.AddFlat(rqJob)...)
	if err != nil {
		fmt.Println("HMSET", err)
	}
}

//This method encodes and then enqueues the jobs in Redis
//TODO: Expose a EnQAndStart method
func (job *RQJob) Enqueue() {
	rqJob := map[string]string{"data": job.EncodeJob()}
	job.EnqueueJob(rqJob)
}

//This method triggers the job in RQ worker
func (job *RQJob) Start() {
	_, err := rqRedisPool.Execute("RPUSH", "rq:queue:default", job.Id)
	if err != nil {
		fmt.Println("RPUSH", err)
	}
}

//This methid returns the result of the job
//TODO: key doesn't exist case; display nil if job is not yet processed
func (job *RQJob) Result() string {
	queueId := job.QueueId()
	result, err := redis.String(rqRedisPool.Execute("HGET", queueId, "result"))
	if err != nil {
		return "nil"
	}
	return DecodeResult(result)
}
