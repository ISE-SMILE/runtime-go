

zip:
	go build -o exec .
	zip main.zip exec
	rm exec

payload: zip
	echo -n "{\"value\":{\"code\":\"" > payload.json
	base64 -w 0 main.zip >> payload.json
	echo -n "\",\"Binary\":true}}" >> payload.json

local: payload
	curl -i -X POST -H "Content-Type: application/json" -d @payload.json http://localhost:8080/init

clean:
	-rm payload.json
	-rm main.zip
	-rm exec