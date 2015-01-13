####Go RQ Client  
Go client to enqueue jobs in RQ.  

Example:

        package main

        import (
                "fmt"
                "github.com/goibibo/t-settings"
                "github.com/goibibo/gorq"
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


Check this gist for RQ internals: https://github.com/vireshas/gorq-rcvr  

You might to use it along with https://github.com/vireshas/gorq-rcvr  

