// __mocks__/Logger.js

const create = jest.fn(() => ({
	info: jest.fn(),
	child: create,
	error: jest.fn(),
}));

module.exports = {
	create,
};
