function login() {
    var login = document.getElementById("loginInput").value;
    var password = document.getElementById("passwordInput").value;

    var credentials = {};

    if (login.includes('@')) {
        credentials = {email: login, password: password}
    } else {
        credentials = {username: login, password: password}
    }

    var encodedCredentials = JSON.stringify(credentials);
    console.log(encodedCredentials);

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (this.readyState === 4) {
            var responseBody = JSON.parse(this.responseText);
            if (this.status === 202) {
                var token = responseBody.body;

                sessionStorage.setItem("token", token);

                window.location.replace("/")
            } else {
                var errorMessage = responseBody.error;

                alert(errorMessage);

                sessionStorage.removeItem("token")
            }
        }
    };

    request.open("POST", "/api/user/login", true);
    request.setRequestHeader("Content-type", "application/json");
    request.send(encodedCredentials);
}
