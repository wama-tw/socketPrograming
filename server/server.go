package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

var name = make(map[string]string)

type Message struct {
	Author string `json:"author"`
	Text   string `json:"text"`
	Time   string `json:"time"`
	Exit   bool   `json:"exit"`
	Join   bool   `json:"join"`
}

func main() {
	service := "localhost:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	fmt.Println(tcpAddr)
	checkError(err)

	var connections []net.Conn
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		connections = append(connections, conn)
		name[conn.RemoteAddr().String()] = randomNameGenerator()
		go handleClient(conn, &connections)
	}
}

func handleClient(conn net.Conn, connections *[]net.Conn) {
	// conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	defer conn.Close() // close connection before exit

	println("connected: ", conn.RemoteAddr())
	sendJoinWaitGroup := &sync.WaitGroup{}
	t := time.Now()
	for _, connection := range *connections {
		sendJoinWaitGroup.Add(1)
		go func(connection net.Conn, sendJoinWaitGroup *sync.WaitGroup) {
			encoder := json.NewEncoder(connection)
			println("sending message to ", fmt.Sprint(connection.RemoteAddr()))
			err := encoder.Encode(Message{
				Author: name[conn.RemoteAddr().String()],
				Time:   timeFormat(t),
				Join:   true,
			})
			if err != nil {
				println("Error occur while sending to", connection.RemoteAddr().String(), ": ", err.Error())
			}
			sendJoinWaitGroup.Done()
		}(connection, sendJoinWaitGroup)
	}
	sendJoinWaitGroup.Wait()

	handleClientWaitGroup := &sync.WaitGroup{}
	handleClientWaitGroup.Add(2)
	var done context.CancelFunc
	ctx, done := context.WithCancel(context.Background())
	message := make(chan Message, 10)
	go getRequest(conn, message, done, handleClientWaitGroup)               // Get request
	go sendResponse(conn, message, ctx, handleClientWaitGroup, connections) // Response

	handleClientWaitGroup.Wait()
	println(conn.RemoteAddr(), "handleClient returned")
	(*connections) = remove((*connections), conn)
	sendExitWaitGroup := &sync.WaitGroup{}
	t = time.Now()
	for _, connection := range *connections {
		sendExitWaitGroup.Add(1)
		go func(connection net.Conn, sendExitWaitGroup *sync.WaitGroup) {
			encoder := json.NewEncoder(connection)
			println("sending message to ", fmt.Sprint(connection.RemoteAddr()))
			err := encoder.Encode(Message{
				Author: name[conn.RemoteAddr().String()],
				Time:   timeFormat(t),
				Exit:   true,
			})
			if err != nil {
				println("Error occur while sending to", connection.RemoteAddr().String(), ": ", err.Error())
			}
			sendExitWaitGroup.Done()
		}(connection, sendExitWaitGroup)
	}
	sendExitWaitGroup.Wait()
	return
}

func getRequest(
	conn net.Conn, message chan Message,
	done context.CancelFunc,
	handleClientWaitGroup *sync.WaitGroup,
) {
	defer handleClientWaitGroup.Done()
	var received Message
	decoder := json.NewDecoder(conn)
	for {
		err := decoder.Decode(&received)
		request := received
		received = Message{} // clear last read content
		request_msg := request.Text
		if err != nil {
			fmt.Println(fmt.Sprint(conn.RemoteAddr())+" Fatal error: ", err.Error())
			request.Exit = true
		}
		if request.Exit {
			message <- Message{
				Author: name[conn.RemoteAddr().String()],
				Exit:   true,
			}
			done()
			return // connection already closed by client
		}
		// respond if not empty
		println(request_msg, "is empty? : ", isEmpty(request_msg))
		if isEmpty(request_msg) {
			continue
		}

		message <- request
	}
}

func sendResponse(
	conn net.Conn,
	message chan Message,
	ctx context.Context,
	handleClientWaitGroup *sync.WaitGroup,
	connections *[]net.Conn,
) {
	defer handleClientWaitGroup.Done()
	for {
		broadcastMsg := <-message
		if !broadcastMsg.Exit {
			println("client sent: ", broadcastMsg.Text)
			sendResponseWaitGroup := &sync.WaitGroup{}
			for _, connection := range *connections {
				sendResponseWaitGroup.Add(1)
				go func(connection net.Conn, sendResponseWaitGroup *sync.WaitGroup) {
					encoder := json.NewEncoder(connection)
					sendingMsg := broadcastMsg
					println("sending message to ", connection.RemoteAddr().String())
					if sendingMsg.Author == connection.RemoteAddr().String() {
						sendingMsg.Author = "我"
					} else {
						sendingMsg.Author = name[conn.RemoteAddr().String()]
					}
					err := encoder.Encode(sendingMsg)
					if err != nil {
						println("Error occur while sending to", connection.RemoteAddr().String(), ": ", err.Error())
					}
					sendResponseWaitGroup.Done()
				}(connection, sendResponseWaitGroup)
			}
			sendResponseWaitGroup.Wait()
			println("Sent to clients: ", broadcastMsg.Text)
			println("-------------------------------------")
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func timeFormat(t time.Time) string {
	formatted := ""
	if t.Hour() < 10 {
		formatted += "0"
	}
	formatted += fmt.Sprint(t.Hour())
	formatted += ":"
	if t.Minute() < 10 {
		formatted += "0"
	}
	formatted += fmt.Sprint(t.Minute())

	return formatted
}

func randomNameGenerator() string {
	names := []string{"海鷗", "鴿子", "鶴", "老鷹", "麻雀", "燕子", "天鵝", "鵝", "啄木鳥", "鸚鵡", "烏鴉", "金絲雀", "紅鶴", "貓", "孔雀", "企鵝", "雞", "火雞", "鴨子", "黑面琵鷺", "美洲豹", "花豹", "雲豹", "石虎", "印度豹", "獵豹", "獅子", "老虎", "獾", "豬", "公豬", "熊", "浣熊", "棕熊", "灰熊", "北極熊", "鬣狗", "海豹", "海象", "海狗", "海獅", "水獺", "水豚", "鹿", "麋鹿", "馴鹿", "梅花鹿", "斑馬", "大象", "長頸鹿", "羚羊", "山羊", "綿羊", "羊駝", "草泥馬", "河馬", "袋鼠", "無尾熊", "天竺鼠", "倉鼠", "老鼠", "松鼠", "公牛", "母牛", "水牛", "犀牛", "小狗", "小貓", "兔子", "野兔", "狼", "雪貂", "穿山甲", "食蟻獸", "狐狸", "狐濛", "猴子", "大猩猩", "黑猩猩", "蝙蝠", "刺蝟", "鴨嘴獸", "土撥鼠", "驢子", "馬", "駱駝", "臭鼬", "熊貓", "馬來膜", "長臂猿", "鱷魚", "短吻鱷魚", "蛇", "眼鏡蛇", "大蟒蛇", "烏龜", "青蛙", "蟾蜍", "變色龍", "壁虎", "蜥蜴", "烏賊", "章魚", "鮪魚", "鮭魚", "牡蠣", "蛤蠣", "金魚", "蝦子", "螃蟹", "龍蝦", "魟魚", "扇貝", "鯨魚", "海馬", "水母", "海豚", "鯊魚", "蝴蝶", "毛毛蟲", "蜜蜂", "蚊子", "蒼蠅", "螞蟻", "蟑螂", "蜈蚣", "蜘蛛", "蠍子"}
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := 127
	return ("匿名" + names[rand.Intn(max-min+1)+min])
}

func remove(s []net.Conn, remove net.Conn) []net.Conn {
	for index, element := range s {
		if element == remove {
			s[index] = s[len(s)-1]
			break
		}
	}
	return s[:len(s)-1]
}

func isEmpty(msg string) bool {
	msg = strings.Replace(msg, " ", "", -1)
	msg = strings.Replace(msg, "\n", "", -1)
	msg = strings.Replace(msg, "\r", "", -1)
	msg = strings.Replace(msg, "\t", "", -1)
	return (msg == string(make([]byte, 128)) || msg == "")
}

func checkError(err error) {
	if err != nil {
		error_msg := fmt.Sprintf("Fatal error: %s", err.Error())
		log.Fatal(error_msg)
	}
}
