const rp = require('request-promise');
const Codefresh = require('./../Codefresh');
const { create: createLogger } = require('./../../services/Logger');

jest.mock('request-promise');
jest.mock('./../../services/Logger');

const getFakeMetadata = () => ({ name: 'unit-test' });

const getFakeConfig = () => ({ baseURL: 'fake-url', token: 'fake-token' });

const createCodefreshAPI = () => new Codefresh(getFakeMetadata(), getFakeConfig());

describe('Codefresh API unit tests', () => {
	describe('Construction', () => {
		it('Should construct', () => {
			expect(createCodefreshAPI).not.toThrowError();
		});

		it('Should set values on this', () => {
			const api = createCodefreshAPI();
			expect(Object.keys(api)).toStrictEqual(['options', 'metadata', 'defaults']);
		});
	});

	describe('Initialization', () => {
		it('Shoudl initialize', () => expect(createCodefreshAPI().init()).resolves.toEqual());
	});

	describe('Calls', () => {
		describe('_call', () => {
			it('Should have the default values', () => {
				const spy = jest.fn();
				rp.__setRequestMock(spy);
				return createCodefreshAPI()
					._call()
					.then(() => {
						expect(spy).toHaveBeenCalledTimes(1);
						expect(spy.mock.calls[0][0]).toMatchObject({
							baseUrl: expect.any(String),
							headers: expect.any(Object),
							json: expect.any(Boolean),
							timeout: expect.any(Number),
						});
					});
			});

			it('Should have default headers', () => {
				const spy = jest.fn();
				rp.__setRequestMock(spy);
				return createCodefreshAPI()
					._call()
					.then(() => {
						expect(spy.mock.calls[0][0]).toHaveProperty('headers.Authorization');
						expect(spy.mock.calls[0][0]).toHaveProperty('headers.Codefresh-Agent-Name');
						expect(spy.mock.calls[0][0]).toHaveProperty('headers.Codefresh-Agent-Version');
					});
			});

			it('Should have default timeout', () => {
				const spy = jest.fn();
				rp.__setRequestMock(spy);
				return createCodefreshAPI()
					._call()
					.then(() => {
						expect(spy.mock.calls[0][0]).toHaveProperty('timeout', 30000);
					});
			});

			it('Should set default request json to be true', () => {
				const spy = jest.fn();
				rp.__setRequestMock(spy);
				return createCodefreshAPI()
					._call()
					.then(() => {
						expect(spy.mock.calls[0][0]).toHaveProperty('json', true);
					});
			});

			it('Should set baseURL', () => {
				const spy = jest.fn();
				rp.__setRequestMock(spy);
				return createCodefreshAPI()
					._call()
					.then(() => {
						expect(spy.mock.calls[0][0]).toHaveProperty('baseUrl', 'fake-url');
					});
			});
		});

		describe('fetchTasksToExecute', () => {
			it('Should set the url', () => {
				const spy = jest.fn();
				rp.__setRequestMock(spy);
				return createCodefreshAPI()
					.fetchTasksToExecute(createLogger(getFakeMetadata()))
					.then(() => {
						expect(spy.mock.calls[0][0]).toHaveProperty('url', '/api/tasks/unit-test');
					});
			});
		});

		describe('reportStatus', () => {
			it('Should call Codefresh API', () => {
				const spy = jest.fn();
				rp.__setRequestMock(spy);
				return createCodefreshAPI()
					.reportStatus(createLogger(getFakeMetadata()), {
						status: {
							message: 'All good',
						},
					})
					.then(() => {
						expect(spy.mock.calls[0][0]).toHaveProperty('url', '/api/runtime-environments/status/unit-test');
						expect(spy.mock.calls[0][0]).toHaveProperty('method', 'PUT');
						expect(spy.mock.calls[0][0]).toHaveProperty('body');
					});
			});
		});

		describe('getLatest', () => {
			it('Should get last version', () => {});

			it('Should retry get last version if a call failed', () => {});
		});
	});
});
