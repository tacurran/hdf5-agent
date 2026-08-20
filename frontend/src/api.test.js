import { describe, expect, it } from 'vitest';
import { datasetUrl, fileUrl, getApiBase, readError } from './api';

describe('api helpers', () => {
  it('uses the versioned API prefix', () => {
    expect(getApiBase()).toBe('/api/v1');
  });

  it('encodes file names in URLs', () => {
    expect(fileUrl('run 1.h5')).toBe('/api/v1/files/run%201.h5');
  });

  it('puts the dataset path in the query string', () => {
    expect(datasetUrl('a.h5', '/measurements/waveform')).toBe(
      '/api/v1/files/a.h5/datasets?path=%2Fmeasurements%2Fwaveform',
    );
  });

  it('reads structured error messages', async () => {
    const response = {
      json: async () => ({ error: { code: 'not_found', message: 'missing file' } }),
      statusText: 'Not Found',
    };
    await expect(readError(response)).resolves.toBe('missing file');
  });
});
