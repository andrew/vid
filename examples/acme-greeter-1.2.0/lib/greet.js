const { execFile } = require('child_process');

module.exports = function greet(name) {
  execFile('echo', ['Hello', name]);
};
