####Go RQ Client


Go client to enqueue jobs in RQ.

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
                job.Result()
        }
