var statusButton = document.getElementById("changeStatusButton");
statusButton.onclick = changeStatus;

var addFriendButton = document.getElementById("addFriendButton");
addFriendButton.onclick = addFriend;

var xhrFriends = new XMLHttpRequest();
var xhrStatus = new XMLHttpRequest();
var xhrAddFriend = new XMLHttpRequest();

getFriends();

function changeStatus() {
	var statusText = document.getElementById("statusInput").value;
	// encodeURIComponent om UTF-8 te gebruiken en speciale karakters om te zetten naar code
	var param = "newStatus=" + encodeURIComponent(statusText);
	xhrStatus.open("POST", "status", true);
	xhrStatus.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	xhrStatus.send(param); 
	document.getElementById("status").textContent = statusText;
}

function addFriend() {
	var friendName = document.getElementById("addFriendInput").value;
	// encodeURIComponent om UTF-8 te gebruiken en speciale karakters om te zetten naar code
	var param = "newFriend=" + encodeURIComponent(friendName);
	xhrAddFriend.open("POST", "addfriend", true);
	xhrAddFriend.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	xhrAddFriend.send(param); 
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
			name.textContent = serverResponse[friend].Name;
			row.appendChild(name);

			var email = document.createElement("td");
			email.id = "email"
			email.textContent = serverResponse[friend].Email;
			row.appendChild(email);

			var stat = document.createElement("td");
			stat.textContent = serverResponse[friend].Status;
			row.appendChild(stat);

			var chatWithTd = document.createElement("td");
			var chatWithButton = document.createElement("button");
			chatWithButton.onclick = chatWithButtonHandler;
			chatWithButton.textContent = "chat";
			chatWithTd.appendChild(chatWithButton);

			if($msgReceiver == email.textContent){
				$(row).css("background-color", "green");
			}
			row.appendChild(chatWithTd);

			friendTable.appendChild(row);
		}

		setTimeout(getFriends, 1000);
		//console.log(friendTable);
	}
}

$msgReceiver = "";

$input = $("#msgInput");
$output = $("#chatWindow");
$msgButton = $("#sendMsgButton");

$(document).ready(function() {
	$msgButton.click(sendMsg);
	receiveMsgs();
});

function chatWithButtonHandler(event){
	email = $(this).parent().parent().find("#email").text();
	$msgReceiver = email;
}

function sendMsg(){
	$.post("sendmsg", {msg: $input.val(), msgreceiver: $msgReceiver});
}

function receiveMsgs(){
	$.get("messages", {msgreceiver: $msgReceiver}, function(data) {
		// $output.text(data);
		msgOutput(data)
	});
	setTimeout(receiveMsgs, 1000);
}

function msgOutput(data) {
	$output.empty();
	messages = JSON.parse(data);
	$(messages).each(function(i, msgObj){
		from = msgObj.From;
		msg = msgObj.Msg;
		$("<p>").html("<span>" + from + ": </span>" + msg).appendTo($output);
	});
}




















