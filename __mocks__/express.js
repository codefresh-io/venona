module.exports = jest.fn().mockImplementation(() => ({
	listen: (port, cb) => cb(),
}));
