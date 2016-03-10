package main

import (
    "fmt"
    "os"
    "os/user"
    "strings"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "bufio"
    "time"
    "strconv"
)

func main() {
  command := "default"

  if len(os.Args) > 1 {
    command = os.Args[1]
  }

  if command == "init" {
    slack_init()
  } else {
    members := fetch_members()
    messages := fetch_messages()

    for i:= range messages {
      fmt.Println(members[messages[i].UserId])
      var timestamp int64
      timestamp, _ = strconv.ParseInt(strings.Split(messages[i].PostedAt, ".")[0], 10, 64)
      fmt.Println(time.Unix(timestamp, 0))

      m := messages[i].Text
      for k, v := range members {
        m = strings.Replace(m, k, v, -1)
      }

      fmt.Println(m)

      fmt.Print("\n")
    }
  }


}

type Channel struct {
  Id string `json:"id"`
  Name string `json:"name"`
}

type Data struct {
  Ok bool
  Channels []Channel
}

type Messages struct {
  Ok bool
  Messages []Message
}

type Message struct {
  UserId string `json:"user"`
  Text string `json:"text"`
  PostedAt string `json:"ts"`
}

type Members struct {
  Members []Member
}

type Member struct {
  Id string `json:"id"`
  Name string `json:"name"`
  Icon string `json:"profile:image_192"`
}

func get_init() (string, string) {
  userFile := get_setting_filename()
  fp, err := os.Open(userFile)
  if err != nil {
    panic(err)
  }
  defer fp.Close()
  reader := bufio.NewReaderSize(fp, 4096)
  line, _, err := reader.ReadLine()
  if err != nil {
    panic(err)
  }
  inits := strings.Split(string(line), ",")
  return inits[0], inits[1]
}

func fetch_messages() []Message{
  token, channel_id := get_init()

  response, _ := http.Get(fmt.Sprintf("https://slack.com/api/channels.history?token=%s&channel=%s&count=10", token, channel_id))
  body, _ := ioutil.ReadAll(response.Body)
  defer response.Body.Close()
  b := []byte(string(body))
  var d Messages
  err := json.Unmarshal(b, &d)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
  }

  return d.Messages
}

func fetch_members() map[string]string {
  token, _ := get_init()

  response, _ := http.Get(fmt.Sprintf("https://slack.com/api/users.list?token=%s", token))
  body, _ := ioutil.ReadAll(response.Body)
  defer response.Body.Close()
  b := []byte(string(body))
  var d Members
  err := json.Unmarshal(b, &d)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
  }

  m := map[string]string{}
  for i:= range d.Members {
    m[d.Members[i].Id] = d.Members[i].Name
  }

  return m
}

func get_setting_filename() string {
  usr, _ := user.Current()
  return strings.Replace("~/.slatic",  "~", usr.HomeDir, 1)
}

func slack_init() {

  // SlackのAPIトークンを入力してもらう
  fmt.Print("[Slack token] > ")
  var token string
  fmt.Scan(&token)

  fmt.Println("[Select channel] ")

  lists := fetch_channels()
  for i:= range lists.Channels {
    fmt.Printf("#%s\n", lists.Channels[i].Name)
  }

  // チャンネルリストから対象のものを選んでもらう
  fmt.Print("please input channel name > ")

  var channel_name string
  fmt.Scan(&channel_name)
  channel_name = strings.Replace(channel_name, "#", "", 1)

  var channel_id string
  for i:= range lists.Channels {
    if channel_name == lists.Channels[i].Name {
      channel_id = lists.Channels[i].Id
      break
    }
  }

  // ユーザのホームディレクトリに、設定ファイルを保存する
  userFile := get_setting_filename()

  fout, err := os.Create(userFile)
  if err != nil {
      fmt.Println(userFile, err)
      return
  }
  defer fout.Close()
  fout.WriteString(fmt.Sprintf("%s,%s", token, channel_id))
}

func fetch_channels() Data {
  token, _ := get_init()

  response, _ := http.Get(fmt.Sprintf("https://slack.com/api/channels.list?token=%s", token))
  body, _ := ioutil.ReadAll(response.Body)
  defer response.Body.Close()
  b := []byte(string(body))
  var d Data
  err := json.Unmarshal(b, &d)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
  }

  return d
}
