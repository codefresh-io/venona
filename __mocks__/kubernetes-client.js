const Client = jest.fn();
const config = {
	getInCluster: jest.fn(),
};

module.exports = {
	Client,
	config,
};
