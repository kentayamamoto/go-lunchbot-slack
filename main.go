package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/mvdan/xurls"
	"github.com/nlopes/slack"
)

func envLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// reffered from http://blog.kaneshin.co/entry/2016/12/03/162653
func run(api *slack.Client) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// テキストから配列を読み込む
	filepath := "data/list.txt"
	lines := fromFile(filepath)

	regexpGoLunch := regexp.MustCompile(`^ランチ(いきたい|行きたい)`)
	regexpAdd := regexp.MustCompile(`^いってきた`)

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				log.Print("LunchBot started!")
				log.Println(lines)

			case *slack.MessageEvent:
				// 都度最新のファイルを読み取る
				lines := fromFile(filepath)
				message := ev.Text
				if regexpGoLunch.MatchString(message) {
					// ランダムにURLを返す
					shuffle(lines)
					rtm.SendMessage(rtm.NewOutgoingMessage(lines[0], ev.Channel))
				} else if regexpAdd.MatchString(message) {
					// URL文字列が存在すれば、リストのファイルに追加する
					urls := xurls.Strict.FindAllString(message, -1)
					if urls == nil {
						continue
					}
					writeNewLine(urls[0], filepath)
					rtm.SendMessage(rtm.NewOutgoingMessage("リストに追加したよ！", ev.Channel))
				}

			case *slack.InvalidAuthEvent:
				log.Print("Invalid credentials")
				return 1
			}
		}
	}
}

func shuffle(data []string) {
	n := len(data)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		data[i], data[j] = data[j], data[i]
	}
}

// referred from http://qiita.com/jpshadowapps/items/ae7274ec0d40882d76b5
func fromFile(filePath string) []string {
	// ファイルを開く
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File %s could not read: %v\n", filePath, err)
		os.Exit(1)
	}

	// 関数return時に閉じる
	defer f.Close()

	// Scannerで読み込む
	// lines := []string{}
	lines := make([]string, 0, 100)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// appendで追加
		lines = append(lines, scanner.Text())
	}
	if serr := scanner.Err(); serr != nil {
		fmt.Fprintf(os.Stderr, "File %s scan error: %v\n", filePath, err)
	}

	return lines
}

func writeNewLine(addtext string, path string) error {
	fileHandle, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	fileHandle.WriteString(addtext + "\n")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	envLoad()
	api := slack.New(os.Getenv("SLACK_API_TOKEN"))
	os.Exit(run(api))
}
