const _ = require('lodash');

let requestMocks;

/**
 * mocks request function
 * __resetRequestMock should be called before each test execution to clean the state
 * not calling __resetRequestMock before each test will
 * cause unexpected behavior and should be avioded
 */
const rp = jest.fn(async (opts) => {
	if (requestMocks) {
		const requestMock = requestMocks.shift();
		return requestMock(opts);
	} if (!requestMocks) {
		throw new Error('no mocked request was set. \nuse rp.__setRequestMock to set the mock or excplicitly unmock with: jest.unmock(@codefresh-io/service-base)');
	} else {
		throw new Error('request was called more than the amount of passed mocks');
	}
});

/**
 * should be called with a function or an array of functions to control
 * the sequence of request behavior function
 * @private
 */
rp.__setRequestMock = (func) => {
	rp.__resetRequestMock();
	const funcArray = _.isArray(func) ? func : [func];
	requestMocks = funcArray;
};

/**
 * should be called before each test
 */
rp.__resetRequestMock = () => {
	rp.mockClear();
	requestMocks = undefined;
};

module.exports = rp;
