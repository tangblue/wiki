package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/websocket"
)

type Data struct {
	MyData string `json:"myData"`
}

func main() {
	http.HandleFunc("/websocket", wsHandler)
	http.Handle("/wsdata", websocket.Handler(wsDataHandler))
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, page)
}

func wsDataHandler(ws *websocket.Conn) {
	dataCh := make(chan Data, 1)
	isRunning := true

	go func() {
		for isRunning {
			dataCh <- Data{strconv.Itoa(rand.Int())}
			time.Sleep(500 * time.Millisecond)
		}
		close(dataCh)
		log.Printf("stop sending data")
	}()

	for data := range dataCh {
		err := websocket.JSON.Send(ws, data)
		if err != nil {
			log.Printf("error sending data: %v\n", err)
			break
		}
	}
	isRunning = false
	log.Printf("ws end")
}

const page = `
<html>
  <head>
      <title>Hello WebSocket</title>

      <script type="text/javascript">
      var sock = null;
      var myData = "";
      function update() {
          var p1 = document.getElementById("my-data-plot");
          p1.innerHTML = myData;
      };
      window.onload = function() {
		  let protocal = location.protocol === "https:" ? "wss:" : "ws:";
          sock = new WebSocket(protocal+"//"+location.host+"/wsdata");
          sock.onmessage = function(event) {
              var data = JSON.parse(event.data);
              myData = data.myData;
              update();
          };
      };
      </script>
  </head>
  <body>
      <div id="header">
          <h1>Hello WebSocket</h1>
      </div>
      <div id="content">
          <div id="my-data-plot"></div>
      </div>
  </body>
</html>
`
