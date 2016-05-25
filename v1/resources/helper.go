// Copyright 2016 Tim Foerster <github@mailserver.1n3t.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
  "runtime"
  "bytes"
//  gzip "github.com/klauspost/pgzip"
  "compress/gzip"
  "io/ioutil"
)

func setCpu() {
  runtime.GOMAXPROCS(runtime.NumCPU())
}

func Zip(str string) string {
  var b bytes.Buffer
  w := gzip.NewWriter(&b)

  //w.SetConcurrency(100000, 10) 
  w.Write([]byte(str))
  w.Close()
  
  return b.String()
}

func Unzip(str string) string {
  read_buf := new(bytes.Buffer)
  read_buf.WriteString(str)

  r, _ := gzip.NewReader(read_buf)
  defer r.Close()

  unzip, _ := ioutil.ReadAll(r)

  return string(unzip)
}
