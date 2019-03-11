var commentButton = document.getElementById("commentButton");
commentButton.onclick = postComment;
var webSocket = new WebSocket("ws://localhost:8080/comment");

function postComment(){
	var comment = {
		text : document.getElementById("comment").value,
		name : document.getElementById("name").value,
		rating : document.getElementById("rating").value
	};
	webSocket.send(JSON.stringify(comment));
}

webSocket.onmessage = function(event){
	var comments = document.getElementById("topic1-comments");
	var row = document.createElement("tr");
	var nameCell = document.createElement("td");
	var ratingCell = document.createElement("td");
	var commentCell = document.createElement("td");

	var comment = JSON.parse(event.data);
	nameCell.textContent = comment.name
	ratingCell.textContent = comment.rating
	commentCell.textContent = comment.text

	row.appendChild(nameCell);
	row.appendChild(ratingCell);
	row.appendChild(commentCell);
	comments.appendChild(row);
};

//webSocket.onopen = function(event){
//	writeResponse("Connection opened")
//};

//webSocket.onclose = function(event){
//	writeResponse("Connection closed");
//};


//function closeSocket(){
//	webSocket.close();
//}

