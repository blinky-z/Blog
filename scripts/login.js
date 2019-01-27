function login() {
    var login = document.getElementById("loginInput").value;
    var password = document.getElementById("passwordInput").value;

    var credentials = {};

    if (login.includes('@')) {
        credentials = {email: login, password: password}
    } else {
        credentials = {login: login, password: password}
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
                window.location.replace("http://localhost:8080/admin")
            } else {
                var errorMessage = responseBody.error;

                console.log(errorMessage);

                sessionStorage.removeItem("token")
            }
        }
    };

    request.open("POST", "http://localhost:8080/user/login", true);
    request.setRequestHeader("Content-type", "application/json");
    request.send(encodedCredentials);
}
