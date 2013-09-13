/*
 *
 * GoStats
 * A simple stats server for Go services
 *
 * (c)2013 Green Man Gaming Limited
 *
 */
package gostats

import (
  "encoding/json"
  "fmt"
  "net/http"
  "regexp"
  "strings"
  "sync"
  "time"

  "github.com/VividCortex/gohistogram"
  "github.com/gorilla/mux"
)

type metric struct {
  histogram *gohistogram.NumericHistogram
  Count     int64
  Sum       int64
  mutex     sync.Mutex
}

func (m *metric) Add(value int64) {
  m.mutex.Lock()
  defer m.mutex.Unlock()

  if m.Count == 0 {
    m.histogram = gohistogram.NewHistogram(200)
  }

  m.Count += 1
  m.Sum += value
  m.histogram.Add(float64(value))
}

func (m *metric) MarshalJSON() ([]byte, error) {
  build_out := make(map[string]interface{})
  build_out["sum"] = m.Sum
  build_out["count"] = m.Count
  build_out["avg"] = m.Sum / m.Count

  percentiles := []float64{0.25, 0.50, 0.75, 0.90, 0.95, 0.99, 0.999, 0.9999}
  re := regexp.MustCompile("(0+)$")
  for _, percentile := range percentiles {
    p := strings.Replace(fmt.Sprintf("p%.2f", percentile*100), ".", "", -1)
    p = re.ReplaceAllString(p, "")
    build_out[p] = int64(m.histogram.Quantile(percentile))
  }

  return json.Marshal(build_out)
}

type StatServe struct {
  Addr         string
  server       *http.Server
  counterMutex sync.Mutex
  counters     map[string]int
  gauges       map[string]int
  labelMutex   sync.Mutex
  labels       map[string]string
  metricMutex  sync.Mutex
  metrics      map[string]*metric
}

func (s *StatServe) FetchStatsFunc() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("content-type", "application/json")

    to_return := make(map[string]interface{})
    to_return["counters"] = s.counters
    to_return["metrics"] = s.metrics
    to_return["labels"] = s.labels
    json_string, _ := json.Marshal(to_return)
    fmt.Fprint(w, string(json_string))
  }
}

func (s *StatServe) configure() {
  // If I'm already configured, there's no point doing it again
  if s.server != nil {
    return
  }

  // Set up the URLs
  r := mux.NewRouter()
  r.HandleFunc("/stats", s.FetchStatsFunc())
  handler := http.NewServeMux()
  handler.Handle("/", r)

  // Set up a server
  s.server = &http.Server{Addr: s.Addr, Handler: handler}
}

// ListenAndServe the Http server - best to use this is as a 'goroutine'
// as this will allow you to run this in the background
func (s *StatServe) ListenAndServe() (e error) {
  s.configure()
  return s.server.ListenAndServe()
}

/*
 * Increment a named counter. We create if it doesn't exist. We also mutex
 * updates. It's best to call this as a goroutine so that you can fire/forget
 */
func (s *StatServe) IncrementCounter(name string) {
  s.counterMutex.Lock()
  defer s.counterMutex.Unlock()

  if _, ok := s.counters[name]; !ok {
    s.counters[name] = 0
  }
  s.counters[name]++
}

/*
 * Time how long it takes to run code in function f. This is designed to work
 * as a wrapper as f takes/returns nothing.
 *
 * So:
 *
 * func foo() int {
 *   return_value = 0
 *
 *   gostats.Stats.Time("foo", func() {
 *     return_value = 1
 *   })
 *
 *   return return_value
 * }
 */
func (s *StatServe) Time(name string, f func()) {
  s.metricMutex.Lock()
  if _, ok := s.metrics[name]; !ok {
    s.metrics[name] = new(metric)
  }
  s.metricMutex.Unlock()

  start := time.Now().UnixNano()
  f()
  end := time.Now().UnixNano()

  m := s.metrics[name]
  m.Add(end - start)
}

/*
 * Set an arbitrary label with a value. Nothing fancy here
 */
func (s *StatServe) Label(name string, value string) {
  s.labelMutex.Lock()
  defer s.labelMutex.Unlock()

  s.labels[name] = value
}

// A default Stats singleton for us
var Stats = &StatServe{
  counters: make(map[string]int),
  metrics:  make(map[string]*metric),
  labels:   make(map[string]string),
}
