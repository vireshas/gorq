####Go RQ Client  
Go client to enqueue jobs in RQ.  

RQ flow tried using python:  
>   
> from functools import partial   
> import cPickle as pickle   
> dumps = partial(pickle.dumps, protocol=pickle.HIGHEST_PROTOCOL)   
>
> Assuming default queue   
> client = redis.Redis()  
> #every job has an enqueue id  
> job_id = "101"  
> #this is the expected format for a func and its params  
> job_tuple = "add.add", None, (100,2), {} # {"pubsub": "true"}  
> #encode with pickle; rq expects pickled value in key "data"  
> job = {  
>    "data" : dumps(job_tuple),  
> }  
> #push the job; ttl of 500  
> client.hmset("rq:job:" + job_id, job)  
> #job is processed as soon as you push the job_id to queue  
> client.rpush("rq:queue:default", job_id)  
> #sleep for a while; once processed data is set in redis  
> time.sleep(2)  
> #result would be available in key "result"  
> result = client.hgetall("rq:job:" + job_id)  
> #decode pickled value before printing  
> print loads(result["result"])  
>  

Example;
        package main

        import (
                "fmt"
                "github.com/goibibo/t-settings"
                "github.com/vireshas/gorq"
                "time"
        )

        func main() {
                settings.Configure()
                gorq.InitRedisPool()
                kwargs := map[string]string{"pubsub": "true"}
                job := gorq.NewRQJob("add.add", []string{"14", "14"}, kwargs)
                job.Enqueue()
                job.Start()
                time.Sleep(2 * time.Second)
                fmt.Println("fetching results")
                fmt.Println(job.Result())
        }
