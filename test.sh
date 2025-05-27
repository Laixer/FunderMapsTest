user_id = "7a015c0a-55ce-4b8e-84b5-784bd336"
user_auth_key = "fmsk.v8qfwf8heu89gkd24pzrzk4cjyt8yqs5"

# General
curl http://localhost:3000/api/app/schiedam -H 'Content-Type: application/json'

# Auth
curl http://localhost:3000/api/auth/signin -d '{"email":"yorick@laixer.com","password":"Yw06b7lgLfFNpoY9QA62MFRfpyUS3AXz"}' -H 'Content-Type: application/json'
curl http://localhost:3000/api/auth/token-refresh -H 'Content-Type: application/json' -H 'Authorization: Bearer fmathCfSD1uORIlwaKLfSDKaHvO96WbBNU0J9brN0ZhF'
curl http://localhost:3000/api/auth/change-password -d '{"current_password":"ABC@12345","new_password":"Yw06b7lgLfFNpoY9QA62MFRfpyUS3AXz"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer X'
curl http://localhost:3000/api/auth/forgot-password -d '{"email":"yorick@laixer.com"}' -H 'Content-Type: application/json'
curl http://localhost:3000/api/auth/reset-password -d '{"reset_key":"ecbfc6b7-a268-4a06-b3a0-02e6dfeef5ad","new_password":"ABC@123"}' -H 'Content-Type: application/json'

# OAuth
curl http://localhost:3000/api/v1/oauth2/token -d "client_id=app-mr6n9ckt&client_secret=app-sk-nxcea4xvemnbbjp4nsm46zynyhqum3jc&grant_type=refresh_token&refresh_token=fmrt85g4e977Q5IHNCjQWgUkp4Y4zohL0k3ESPHoLNY8"
curl http://localhost:3000/api/v1/oauth2/token -d "client_id=app-mr6n9ckt&client_secret=app-sk-nxcea4xvemnbbjp4nsm46zynyhqum3jc&grant_type=refresh_token&refresh_token=fmrt85g4e977Q5IHNCjQWgUkp4Y4zohL0k3ESPHoLNY8"
curl http://localhost:3000/api/v1/oauth2/token -d "client_id=app-mr6n9ckt&client_secret=app-sk-nxcea4xvemnbbjp4nsm46zynyhqum3jc&grant_type=authorization_code&code=sdjfnsdjkgn"
curl http://localhost:3000/api/v1/oauth2/token -d "client_id=app-mr6n9ckt&client_secret=app-sk-nxcea4xvemnbbjp4nsm46zynyhqum3jc&grant_type=client_credentials"
curl http://localhost:3000/api/v1/userinfo -H 'Content-Type: application/json' -H 'Authorization: Bearer fmataOSCWcCRbTTqdvdZIJiufhE9vJpfCsPOt57CVEcT'


# https://analytics.fundermaps.com/login/generic_oauth?code=fye3gh487gfierfghi&state=PFlLqj3NfOt2sic3BAt5eXY40UE8gcQMiamqj-1B3L4%3D
# https://api.fundermaps.com/api/v1/oauth2/authorize?client_id=app-8ejdxrfh&redirect_uri=https%3A%2F%2Fanalytics.fundermaps.com%2Flogin%2Fgeneric_oauth&response_type=code&scope=user%3Aemail&state=PFlLqj3NfOt2sic3BAt5eXY40UE8gcQMiamqj-1B3L4%3D

# User
curl http://localhost:3000/api/user/me -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatIvAHiVnfOBCi62rVQ7uHi9zBhSobHzOyLYtqjs1r'
curl http://localhost:3000/api/user/me -H 'Content-Type: application/json' -H 'X-API-Key: fmsk.v8qfwf8heu89gkd24pzrzk4cjyt8yqs5'
curl http://localhost:3000/api/user/metadata -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatIvAHiVnfOBCi62rVQ7uHi9zBhSobHzOyLYtqjs1r'
curl -X PUT http://localhost:3000/api/user/metadata -d '{"metadata":{"lastRotation":"-36.194","lastZoomLevel":"16.72","lastPitchDegree":"38.05","lastCenterPosition":{"lng":4.3714,"lat":51.9}}}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatj3Vh2cQ55vKg90tnOTGGTNm1jNOtWKdvL25Qh0rY'
curl -X PUT http://localhost:3000/api/user/me -d '{"job_title":"Kaas"}' -H 'Content-Type: application/json' -H 'X-API-Key: fmsk.v8qfwf8heu89gkd24pzrzk4cjyt8yqs5'

# Test
# curl -X POST http://localhost:3000/api/test/mail -H 'Content-Type: application/json'

# Management org:b878e0df-59df-4e16-b865-f43a02cca31b user:7c7deb59-b672-4d21-8bf1-9311b60c5cd2 remote_url:https://goldfish-app-4m6mn.ondigitalocean.app/
# Become admin: curl http://localhost:3000/api/auth/signin -d '{"email":"admin@fundermaps.com","password":"ni6DUBZnlqa7S0jgFIwXicMlqOTMRxvG"}' -H 'Content-Type: application/json'
curl http://localhost:3000/api/v1/management/app -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatckau5jV5LCcZqQQgaPDWglK2kML7YOXD3EtMobeV'
curl http://localhost:3000/api/v1/management/org -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatqVgxJUanq0AiCuThpxCYEwpOW5JJQEBSmQpVOtoA'
curl http://localhost:3000/api/v1/management/mapset -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatckau5jV5LCcZqQQgaPDWglK2kML7YOXD3EtMobeV'
curl http://localhost:3000/api/v1/management/mapset/cknycxq5h1f9a17pj578xieqj -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatckau5jV5LCcZqQQgaPDWglK2kML7YOXD3EtMobeV'
curl -X GET     http://localhost:3000/api/v1/management/user -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatckau5jV5LCcZqQQgaPDWglK2kML7YOXD3EtMobeV'
curl -X POST    http://localhost:3000/api/v1/management/user -d '{"email":"heikemavanderkloet@elkien.nl", "password":"x3sMyk6KPzsmczV"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatckau5jV5LCcZqQQgaPDWglK2kML7YOXD3EtMobeV'
curl -X GET     http://localhost:3000/api/v1/management/user/32020859-bdc7-4d35-ae64-6d8ec669cd55 -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatckau5jV5LCcZqQQgaPDWglK2kML7YOXD3EtMobeV'
curl -X POST    http://localhost:3000/api/v1/management/user/32020859-bdc7-4d35-ae64-6d8ec669cd55/reset-password -d '{"password":"blub12"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatBkDYQro9PNvBKECSnpXLTXSMJIk7RVpRQC7InUQ6'
curl -X GET     http://localhost:3000/api/v1/management/user/31c761a4-38e5-441a-b46b-5e2dd69af79b/api-key -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatQVPmQs0HhBAX3957G2WgXOY3R6aZNO94hSFYTr8V'
curl -X GET     http://localhost:3000/api/v1/management/org -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatQVPmQs0HhBAX3957G2WgXOY3R6aZNO94hSFYTr8V'
curl -X POST    http://localhost:3000/api/v1/management/org -d '{"name":"Obvion"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatQVPmQs0HhBAX3957G2WgXOY3R6aZNO94hSFYTr8V'
curl -X GET     http://localhost:3000/api/v1/management/org/b878e0df-59df-4e16-b865-f43a02cca31b -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatQVPmQs0HhBAX3957G2WgXOY3R6aZNO94hSFYTr8V'
curl -X POST    http://localhost:3000/api/v1/management/org/b878e0df-59df-4e16-b865-f43a02cca31b/user -d '{"user_id":"7c7deb59-b672-4d21-8bf1-9311b60c5cd2", "role":"reader"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatQVPmQs0HhBAX3957G2WgXOY3R6aZNO94hSFYTr8V'
curl -X DELETE  http://localhost:3000/api/v1/management/org/d8c19418-c832-4c91-8993-84b8ed641448/user -d '{"user_id":"32020859-bdc7-4d35-ae64-6d8ec669cd55"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatLE99OUywsWRSfU71GSwHtbHFWYRvZZQiRz9zZtx7'
curl -X POST    http://localhost:3000/api/v1/management/org/b878e0df-59df-4e16-b865-f43a02cca31b/mapset -d '{"mapset_id":"7c7deb59-b672-4d21-8bf1-9311b60c5cd2"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatQVPmQs0HhBAX3957G2WgXOY3R6aZNO94hSFYTr8V'
curl -X DELETE  http://localhost:3000/api/v1/management/org/d8c19418-c832-4c91-8993-84b8ed641448/mapset -d '{"user_id":"32020859-bdc7-4d35-ae64-6d8ec669cd55"}' -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatLE99OUywsWRSfU71GSwHtbHFWYRvZZQiRz9zZtx7'

# Data
curl http://localhost:3000/api/data/contractor -H 'Content-Type: application/json' -H 'Authorization: Bearer fmateCnXoNfQWokLtmebsmyBX10B6eiLH8FlBIZAosbd'

# Incident
curl http://localhost:3000/api/incident -d '{"building":"NL.IMBAG.PAND.0497100000004928", "chained_building": true, "foundation_type":"wood", "foundation_damage_characteristics":["crack"]}' -H 'Content-Type: application/json'
curl http://localhost:3000/api/incident/upload -F "files=@dummy.pdf"

# Geocoder
curl http://localhost:3000/api/geocoder/NL.IMBAG.PAND.0606100000004452
curl http://localhost:3000/api/geocoder/NL.IMBAG.PAND.1676100000430461 -H 'Content-Type: application/json'
curl http://localhost:3000/api/geocoder/1952100000002997 -H 'Content-Type: application/json'
curl http://localhost:3000/api/geocoder/NL.IMBAG.NUMMERAANDUIDING.0202200000386458 -H 'Content-Type: application/json'
curl http://localhost:3000/api/geocoder/0202200000386458 -H 'Content-Type: application/json'
curl http://localhost:3000/api/geocoder/NL.IMBAG.PAND.0311100000005965/address -H 'Content-Type: application/json'

# Product
curl http://localhost:3000/api/product/NL.IMBAG.PAND.0606100000004452/analysis -H 'Content-Type: application' -H 'Authorization: Bearer fmatScYW1ULnLvDh2oVLPwLgswrxiFJALOFAL2x1CfHF'
curl http://localhost:3000/api/product/NL.IMBAG.PAND.0355100000725405/statistics -H 'Content-Type: application' -H 'Authorization: Bearer fmatUUVs2gmus2CzoxpcvI5y3znENFOZbqccHdN6unDe'
# curl http://localhost:3000/api/product/NL.IMBAG.PAND.1699100000000456/subsidence -H 'Content-Type: application' -H 'Authorization: Bearer fmatUUVs2gmus2CzoxpcvI5y3znENFOZbqccHdN6unDe'
curl http://localhost:3000/api/product/NL.IMBAG.PAND.1699100000000456/subsidence/historic -H 'Content-Type: application' -H 'Authorization: Bearer fmatUUVs2gmus2CzoxpcvI5y3znENFOZbqccHdN6unDe'

# Mapset
curl http://localhost:3000/api/mapset -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatIvAHiVnfOBCi62rVQ7uHi9zBhSobHzOyLYtqjs1r'
curl http://localhost:3000/api/mapset/clv9n7mz300qe01oc49a05at0 -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatnrAgrL1WMnFtlSMTNxQ7c7J3tHUcSfaWyHQtSZCI'

# Report
curl http://localhost:3000/api/report/NL.IMBAG.PAND.0796100000230874 -H 'Content-Type: application' -H 'Authorization: Bearer fmatPopiEtwtj36kgLp3BggO7msHqCHgwC2dFcMzO49n'

- Fetch org + role
- Always return a JSON response
- GORM: not found is not an error, no need for a stack trace
- Return 404 if not found
- Return 400 + parser error if BodyParser fails
- Combine BodyParser + Viper
- Input validation via Viper
- Return unmarshalled JSON for JSONB data types
- For GetAllX() methods, add limit and offset
- Always order lists


curl https://goldfish-app-4m6mn.ondigitalocean.app/api/auth/signin -d '{"email":"yorick@laixer.com","password":"ABC@123"}' -H 'Content-Type: application/json'
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/user/me -H 'Content-Type: application/json' -H 'Authorization: Bearer fmatVpnfx0AhUkNmhpjzoxV8f4yjfGayt0bNGPNIQCwk'
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/v1/oauth2/token -d "client_id=app-mr6n9ckt&client_secret=app-sk-nxcea4xvemnbbjp4nsm46zynyhqum3jc&grant_type=client_credentials"
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/v4/product/NL.IMBAG.PAND.0355100000725405/analysis -H 'Content-Type: application/json' -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzQyNjAwMTAsImlkIjoiN2EwMTVjMGEtNTVjZS00YjhlLTg0YjUtNzg0YmQzMzYzZDViIn0.gQdE95n7Rk_02VYcNbGvKnfWkRjYNREH1zLmeYgt-7U'
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/v4/product/NL.IMBAG.PAND.1699100000000456/subsidence -H 'Content-Type: application' -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzQyNjAwMTAsImlkIjoiN2EwMTVjMGEtNTVjZS00YjhlLTg0YjUtNzg0YmQzMzYzZDViIn0.gQdE95n7Rk_02VYcNbGvKnfWkRjYNREH1zLmeYgt-7U'
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/v4/product/NL.IMBAG.PAND.1699100000000456/subsidence_historic -H 'Content-Type: application' -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzQyNjAwMTAsImlkIjoiN2EwMTVjMGEtNTVjZS00YjhlLTg0YjUtNzg0YmQzMzYzZDViIn0.gQdE95n7Rk_02VYcNbGvKnfWkRjYNREH1zLmeYgt-7U'
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/geocoder/NL.IMBAG.NUMMERAANDUIDING.0202200000387016 -H 'Content-Type: application/json'
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/geocoder/NL.IMBAG.PAND.0363100012132938 -H 'Content-Type: application/json'
curl https://goldfish-app-4m6mn.ondigitalocean.app/api/geocoder/NL.IMBAG.PAND.0311100000005965/address -H 'Content-Type: application/json'
