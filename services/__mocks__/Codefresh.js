// __mocks__/Codefresh.js
const Codefresh = require.requireActual('./../Codefresh');

module.exports = jest.fn().mockImplementation(() => Codefresh);
