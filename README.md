# GoStats
> Simple Stats for Go Services

At Green Man Gaming we have a large number of services that run and we have a
requirement to monitor them all.

To do this we rely on have stats we can cURL out. We have this via
[Ostrich](https://github.com/twitter/ostrich) from the guys over at
[Twitter](https://twitter.com/twitteross). However, for Google Go we needed something
similar and couldn't find something suitable.

GoStats was born.

    go get github.com/greenmangaming/gostats

Three types of metrics are supported:

 1. Counters
 2. Labels
 3. Timers (count, sum, averages and percentiles)

## Usage

Let's take the default example from
[Writing Web Applications](http://golang.org/doc/articles/wiki/) of a very simple
Go Http Server:

    package main

    import (
      "fmt"
      "net/http"
    )

    func handler(w http.ResponseWriter, r *http.Request) {
      fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
    }

    func main() {
      http.HandleFunc("/", handler)
      http.ListenAndServe(":8080", nil)
    }

Let's do the same things with Stats:

    package main

    import (
      "fmt"
      "net/http"

      "github.com/greenmangaming/gostats"
    )

    func handler(w http.ResponseWriter, r *http.Request) {
      gostats.Stats.Time("root", func() {
        fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
      })
    }

    func main() {
      gostats.Stats.Addr = ":8081"
      go gostats.Stats.ListenAndServe()

      http.HandleFunc("/", handler)
      http.ListenAndServe(":8080", nil)
    }

That's it!

Compile/run and do curl http://localhost:8080/Lee and then
curl http://localhost:8081/stats:

    ➜  ~  curl http://localhost:8080/Lee
    Hi there, I love Lee!%
    ➜  ~  curl http://localhost:8081/stats
    {"counters":{},"labels":{},"metrics":{"root":{"avg":19959,"count":1,"p25":19959,
    "p5":19959,"p75":19959,"p9":19959,"p95":19959,"p99":19959,"p999":19959,"p9999":19959,
    "sum":19959}}}%

Timings are reported in Nanoseconds.

### Counters

Counters are just that. They count every call. I recommend using this as a goroutine:

    go gostats.Stats.IncrementCounter("gophers")

### Labels

Labels are very basic and just allow you to set the current status of something:

    go gostats.Stats.Label("groundhog_day", "yes")

### Timers

Timers are a little bit more involved than the other types. An example of the usage of
it is above.
