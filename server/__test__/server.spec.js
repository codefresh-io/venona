const express = require('express');
const { Server } = require('./../');
const { create: createLogger } = require('./../../services/Logger');
const utils = require('./../../utils');

jest.mock('./../../services/Logger');


const getFakeMetadata = () => ({ name: 'unit-test' });

const getServerConfig = () => ({
	metadata: getFakeMetadata(),
	opt: {
		port: '9000',
	},
	logger: createLogger(getFakeMetadata()),
});

const createServer = () => {
	const config = getServerConfig();
	return new Server(config.metadata, config.opt, config.logger);
};

describe('Server unit testing', () => {
	describe('Constructing new server instace', () => {
		it('Should construct', () => {
			expect(createServer()).toBeInstanceOf(Server);
		});

		it('Should call to looger with message', () => {
			const spy = jest.fn();
			createLogger.mockImplementationOnce(() => ({
				info: spy,
			}));
			createServer();
			expect(spy).toHaveBeenCalled();
			expect(spy).toHaveBeenCalledWith('Starting server component');
		});

		it('Should ensure port was given to the constructor', () => {
			jest.mock('./../../utils');
			jest.spyOn(utils, 'getPropertyOrError');
			createServer();
			expect(utils.getPropertyOrError).toHaveBeenCalled();
		});
	});

	describe('Initialize server', () => {
		it('Should start listening on given port', () => {
			const spy = jest.fn((port, cb) => cb());
			express.mockImplementationOnce(() => ({
				listen: spy,
			}));
			return createServer().init()
				.then(() => {
					const firstArgumentToSpyFunc = spy.mock.calls[0][0];
					expect(firstArgumentToSpyFunc).toEqual('9000');
				});
		});

		it('Should print message after server succcessfully started', () => {
			const spy = jest.fn();
			createLogger.mockImplementationOnce(() => ({
				info: spy,
			}));
			return createServer().init()
				.then(() => {
					expect(spy).toHaveBeenCalled();
					expect(spy).toHaveBeenCalledWith('Starting server component');
				});
		});

		it('Should throw an error when server failed during initialization process', () => {
			const fakeErrorMessage = 'Error!';
			express.mockImplementationOnce(() => ({
				listen: (port, cb) => cb(new Error(`${fakeErrorMessage}`)),
			}));
			const server = createServer();
			return expect(server.init()).rejects.toThrowError(`Failed during server initialization with message: ${fakeErrorMessage}`);
		});
	});
});
