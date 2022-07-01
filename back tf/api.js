/// Comunicacion con Front

var express = require('express')
var app = express();
var bodyParser = require('body-parser')
var jsonParser = bodyParser.json()
const PUERTO = 3000;
var response
var data_json

const net = require('net')

const options = {
    port: 8000,
    host: '127.0.0.1'
}

const client = net.createConnection(options)
app.use(bodyParser.urlencoded({ extended: true }));
app.use(function (req, res, next) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
    next();
  });
app.listen(PUERTO, function () {
    console.log('Servidor https corriendo en puerto ' + PUERTO)
});

app.get('/', function (req, res) {
  
    console.log('Se recibio una peticion GET')

    frame = {
        cmd: "api_get_blockchain",
        sender: "localhost:3000",
        data: [],
    };
    data_eviar = JSON.stringify(frame)
    client.write(data_eviar)
    response=res

});

app.post('/new', jsonParser, (req, res) => {
    data_eviar = {
        cmd: "api_new_block",
        sender: "localhost:3000",
        data: [JSON.stringify(req.body)],
      };
    data_eviar = JSON.stringify(data_eviar)
    console.log(data_eviar)
    client.write(data_eviar)
    res.send("response")

});


/// Comunicacion con Go

client.on("connect", () => {
    console.log(`Conexion a TCP: ${options.port})`);
    data_eviar = {
      cmd: "api_connect",
      sender: `localhost:${PUERTO}`,
      data: [],
    };
    data_eviar = JSON.stringify(data_eviar);
    console.log(data_eviar);
  
    //  connect();
    client.write(data_eviar);
  });

client.on('data', (data) => {
    console.log('el servidor dice:' + data)
    data_json = JSON.parse(data)
    console.log('parceado json:' + String(data_json))
    response.json(data_json)
})

client.on('error', (err) => {
    console.log(err.message)
})

