import axios from "axios";

export function getApiBaseURL(): string {

  let host = process.env.REACT_APP_NOTICEME_API_HOST;

  if (!host) {
    host = window.location.hostname;
  }

  return `${window.location.protocol}//${host}:${process.env.REACT_APP_NOTICEME_API_PORT}/api`;
}

export function getWebsocketBaseURL() {
  let host = process.env.REACT_APP_NOTICEME_API_HOST;

  if (!host) {
    host = window.location.hostname;
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

  return `${protocol}//${host}:${process.env.REACT_APP_NOTICEME_API_PORT}`;
}

export function api() {
  return axios.create({
    baseURL: getApiBaseURL(),
  });
}