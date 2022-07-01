const net =  require('net')
const readline = require('readline-sync')

const options = {
    port: 8000,
    host: '127.0.0.1'
}

const client = net.createConnection(options)

client.on('connect',()=>{
    console.log('Conectado')
    sendLine()
})

client.on('data', (data)=>{
    console.log('el servidor dice:' + data)
    sendLine()
})

client.on('error', (err)=>{
    console.log(err.message)
})
function sendLine(){
    var line = readline.question('\n escriba algo: ')
    if(line == "0"){
        client.end()
    } else {
        client.write(line)
    }
}