var statusButton = document.getElementById("changeStatusButton");
statusButton.onclick = changeStatus;

var xhrFriends = new XMLHttpRequest();
var xhrStatus = new XMLHttpRequest();

getFriends();

function changeStatus() {
	var statusText = getStatusInput();
	// encodeURIComponent om UTF-8 te gebruiken en speciale karakters om te zetten naar code
	var param = "newStatus=" + encodeURIComponent(statusText);
	xhrStatus.open("POST", "status", true);
	xhrStatus.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	xhrStatus.send(param); 
	document.getElementById("status").textContent = statusText;
}

function getStatusInput(){
	return document.getElementById("statusInput").value;
}

function getFriends()
{
	var url = "friends";
	xhrFriends.open("GET", url);
	xhrFriends.onreadystatechange = updateFriendsWhenReady
	xhrFriends.send(null);
}

function updateFriendsWhenReady () {
	if (xhrFriends.status == 200 && xhrFriends.readyState == 4) {
		var serverResponse = JSON.parse(xhrFriends.responseText);
		var friendTable = document.getElementById("friends");
		while (friendTable.firstChild) {
			friendTable.removeChild(friendTable.firstChild);
		}
		for(friend in serverResponse){
			var row = document.createElement("tr");
			var name = document.createElement("td");
			name.textContent = friend;
			row.appendChild(name);
			var stat = document.createElement("td");
			stat.textContent = serverResponse[friend];
			row.appendChild(stat);
			friendTable.appendChild(row);
		}

		setTimeout(getFriends, 1000);
		//console.log(friendTable);
	}
}

// ws code
//var input = document.getElementById("msgInput");
//var output = document.getElementById("chatWindow");
//var msgButton = document.getElementById("sendMsgButton");
//msgButton.onclick = send;
//var socket = new WebSocket("ws://localhost:8080/echo");
//
//socket.onopen = function () {
//	output.innerHTML += "Status: Connected\n";
//};
//
//socket.onmessage = function (e) {
//	output.innerHTML += "Server: " + e.data + "\n";
//};
//
//function send() {
//	socket.send(input.value);
//	input.value = "";
//}
