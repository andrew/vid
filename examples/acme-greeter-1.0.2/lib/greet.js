const { exec } = require('child_process');

module.exports = function greet(name) {
    // say hello to the user
    exec('echo Hello ' + name);
};
