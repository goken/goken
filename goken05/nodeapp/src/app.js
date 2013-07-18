var PORT = 8080;
var server = require('http').createServer(function(req, res){
  res.send('Hello World\n');
});
server.listen(PORT)
console.log('Running on http://localhost:' + PORT);

