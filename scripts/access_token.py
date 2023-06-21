import requests

api_key = 'APIKEY'
email = 'jw1@uvic.ca'
password = 'mysecurepw$'

url = f'https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key={api_key}'
headers = {'Content-Type': 'application/json'}
data = {
    'email': email,
    'password': password,
    'returnSecureToken': True
}

response = requests.post(url, headers=headers, json=data)
response_data = response.json()

if 'idToken' in response_data:
    print(response_data['idToken'])
else:
    print('Login failed. Check your credentials.')
