const API_BASE = (import.meta.env.VITE_API_URL || '/api/v1').replace(/\/$/, '');

/** Base URL of the versioned HDF5 Agent API. */
export function getApiBase() {
  return API_BASE;
}

/** URL for a file resource. */
export function fileUrl(name) {
  return `${API_BASE}/files/${encodeURIComponent(name)}`;
}

/** URL for reading a dataset. */
export function datasetUrl(name, path) {
  const query = new URLSearchParams({ path });
  return `${fileUrl(name)}/datasets?${query.toString()}`;
}

/** Extract a human-readable message from a structured API error. */
export async function readError(response) {
  try {
    const body = await response.json();
    if (body?.error?.message) {
      return body.error.message;
    }
  } catch {
    // ignore parse failures
  }
  return response.statusText || 'Request failed';
}

async function getJSON(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(await readError(response));
  }
  return response.json();
}

/** List HDF5 files in the agent data directory. */
export async function listFiles() {
  const body = await getJSON(`${API_BASE}/files`);
  return body.files || [];
}

/** Load the group/dataset tree for a file. */
export async function getFileStructure(name) {
  return getJSON(fileUrl(name));
}

/** Read dataset values. */
export async function readDataset(name, path) {
  return getJSON(datasetUrl(name, path));
}

/** Create an empty HDF5 file. */
export async function createFile(name) {
  const response = await fetch(`${API_BASE}/files`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  });
  if (!response.ok) {
    throw new Error(await readError(response));
  }
  return response.json();
}

/** Delete an HDF5 file. */
export async function deleteFile(name) {
  const response = await fetch(fileUrl(name), { method: 'DELETE' });
  if (!response.ok) {
    throw new Error(await readError(response));
  }
}

/** Update flattened dataset indices. */
export async function updateDataset(name, path, index, value) {
  const response = await fetch(`${fileUrl(name)}/datasets`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      path,
      indices: [index],
      values: [value],
    }),
  });
  if (!response.ok) {
    throw new Error(await readError(response));
  }
}
