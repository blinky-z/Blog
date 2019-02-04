function registerAccount() {
    var email = document.getElementById("emailInput").value;
    var username = document.getElementById("loginInput").value;
    var password = document.getElementById("passwordInput").value;

    var credentials = {email: email, username: username, password: password};

    var encodedCredentials = JSON.stringify(credentials);
    console.log(encodedCredentials);

    var request = new XMLHttpRequest();
    request.onreadystatechange = function () {
        if (this.readyState === 4) {
            if (this.status === 200) {
                window.location.replace("/login")
            } else {
                var responseBody = JSON.parse(this.responseText);
                var errorMessage = responseBody.error;

                alert(errorMessage);
            }
        }
    };

    request.open("POST", "/api/user/register", true);
    request.setRequestHeader("Content-type", "application/json");
    request.send(encodedCredentials);
}
