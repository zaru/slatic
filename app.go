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

  if "init" == command {
    slack_init()
  } else {
    output_message()
  }


}

// チャンネル一覧
type Channels struct {
  Ok bool
  Channels []Channel
}
type Channel struct {
  Id string `json:"id"`
  Name string `json:"name"`
}

// メッセージ一覧
type Messages struct {
  Ok bool
  Messages []Message
}
type Message struct {
  UserId string `json:"user"`
  Text string `json:"text"`
  PostedAt string `json:"ts"`
}

// JSON出力用
type ParseMessage struct {
  UserName string `json:"user_name"`
  UserIcon string `json:"user_icon"`
  Text string `json:"text"`
  PostedAt time.Time `json:"posted_at"`
}

// メンバー一覧
type Members struct {
  Members []Member
}
type Member struct {
  Id string `json:"id"`
  Name string `json:"name"`
  Profile Profile
}
type Profile struct {
  Icon string `json:"image_192"`
}

// 初期設定
func slack_init() {

  // SlackのAPIトークンを入力してもらう
  fmt.Print("[Slack token] > ")
  var token string
  fmt.Scan(&token)

  fmt.Println("[Select channel] ")

  lists := fetch_channels(token)
  for i:= range lists {
    fmt.Printf("#%s\n", lists[i].Name)
  }

  // チャンネルリストから対象のものを選んでもらう
  fmt.Print("please input channel name > ")

  var channel_name string
  fmt.Scan(&channel_name)
  channel_name = strings.Replace(channel_name, "#", "", 1)

  var channel_id string
  for i:= range lists {
    if channel_name == lists[i].Name {
      channel_id = lists[i].Id
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

// JSONファイルを出力
func output_message() {
  members_name, members_icon := fetch_members()
  messages := fetch_messages()

  parse_messages := []ParseMessage{}

  for i:= range messages {

    // 投稿日時をフォーマット化
    var timestamp int64
    timestamp, _ = strconv.ParseInt(strings.Split(messages[i].PostedAt, ".")[0], 10, 64)

    p := ParseMessage{}
    p.UserName = members_name[messages[i].UserId]
    p.UserIcon = members_icon[messages[i].UserId]
    p.PostedAt = time.Unix(timestamp, 0)
    // mentionのIDをユーザ名に置換
    p.Text = messages[i].Text
    for k, v := range members_name {
      p.Text = strings.Replace(p.Text, k, v, -1)
    }
    parse_messages = append(parse_messages, p)
  }

  // JSONにフォーマット
  bytes, err := json.Marshal(parse_messages)
  if err != nil {
      panic(err)
  }

  // JSONファイルとして出力
  json_file := get_yesterday() + ".json"
  fout, err := os.Create(json_file)
  if err != nil {
    panic(err)
  }
  defer fout.Close()
  fout.WriteString(string(bytes))
}

func fetch_channels(token string) []Channel {

  response, _ := http.Get(fmt.Sprintf("https://slack.com/api/channels.list?token=%s", token))
  body, _ := ioutil.ReadAll(response.Body)
  defer response.Body.Close()
  b := []byte(string(body))
  var d Channels
  err := json.Unmarshal(b, &d)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
  }

  return d.Channels
}

func fetch_members() (map[string]string, map[string]string) {
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

  mem := map[string]string{}
  for i:= range d.Members {
    mem[d.Members[i].Id] = d.Members[i].Name
  }

  icon := map[string]string{}
  for i:= range d.Members {
    icon[d.Members[i].Id] = d.Members[i].Profile.Icon
  }

  return mem, icon
}

func fetch_messages() []Message{
  token, channel_id := get_init()
  oldest, latest := get_date_range()

  uri := fmt.Sprintf("https://slack.com/api/channels.history?token=%s&channel=%s&oldest=%d&latest=%d&count=1000", token, channel_id, oldest, latest)
  response, _ := http.Get(uri)
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

// トークン・設定チャンネルIDを取得
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

// 設定ファイル取得
func get_setting_filename() string {
  usr, _ := user.Current()
  return strings.Replace("~/.slatic",  "~", usr.HomeDir, 1)
}

// ファイル名で使用する日付を取得
func get_yesterday() string {
  t := time.Now()
  t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
  t = t.AddDate(0, 0, -1)
  return fmt.Sprintf("%d-%d-%d", t.Year(), t.Month(), t.Day())
}

// 取得するメッセージの日時レンジを取得
func get_date_range() (int64, int64) {
  today := time.Now()
  today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
  yesterday := today.AddDate(0, 0, -1)
  return yesterday.Unix(), today.Unix()
}
