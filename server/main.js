var input = document.getElementById("input");
var output = document.getElementById("output");

var socket = new WebSocket("ws://localhost:8080/ws");

socket.onopen = function () {
  output.innerHTML += "Status: Connected\n";
};

socket.onmessage = function (e) {
  output.innerHTML += "Server: " + e.data + "\n";
};

function send() {
  socket.send(input.value);
  input.value = "";
}

var canvas = document.getElementById("canvas");
var context = canvas.getContext("2d");
if (context) {
  context.fillRect(0, 0, 150, 100);
}
canvas.addEventListener(
  "mousedown",
  (e) => {
    console.log(e);
  },
  false
);
