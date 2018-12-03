// __mocks__/index.js
const { Server } = require.requireActual('./..');

module.exports = jest.fn().mockImplementation(() => Server);
