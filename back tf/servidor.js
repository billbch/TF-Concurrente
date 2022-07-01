const net = require('net')

const server = net.createServer()

server.on('connection', (socket)=>{
    socket.on('data', (data)=>{
        console.log("mensaje recibido del servidor: "+data)
        socket.write("recibido")
    })
    socket.on('close',()=>{
        console.log("comunicacion finalizada")
    })
    socket.on('error',(err)=>{
        console.log(err.message)
    })
})
server.listen(4000, ()=>{
    console.log('servidor escucha desde: ',server.address().port)
})