let ws; // WebSocket connection
let conn_err; // Div containing elements that should be displayed on WebSocket connection error
let content; // Div for elements to be displayed when WebSocket connection is active (website works like intended)

function loaded() {
  conn_err = document.getElementById("conn_err");
  content = document.getElementById("content");

  connect();
}

function connect() {
  ws = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port);

  ws.addEventListener("open", () => {
    console.log("WebSocket connection established!");
    conn_err.hidden = true;
    content.hidden = false;

    ws.send("Hello?");
  })

  ws.addEventListener("close", () => {
    console.log("WebSocket connection closed!");
    conn_err.hidden = false;
    content.hidden = true;
  });

  ws.addEventListener("error", (err) => {
    console.error("Socket encountered error: ", err.message);
    ws.close();
  });

  ws.addEventListener("message", (e) => {
    parse_message(e.data);
  });

}

window.addEventListener("load", loaded);

function clear_content() {
  content.innerHTML = "";
}

function parse_message(data) {
  if (data == "PONG") {
    console.log(data);
    return
  }

  let msg = JSON.parse(data);
  console.log(msg);

  // Clear previous child nodes
  clear_content();

  let text = document.createElement("h1");
  text.appendChild(document.createTextNode(msg.message));
  content.appendChild(text);
}