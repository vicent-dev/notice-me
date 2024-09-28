import {toast} from "material-react-toastify";
import {getWebsocketBaseURL} from "@/pages/util/api";

export default function connectNoticeMe(clientId: string, clientGroupId: null|string = null): boolean {
  if(!clientGroupId){
    clientGroupId='*';
  }

  const ws = new WebSocket(`${getWebsocketBaseURL()}/ws?id=${clientId}&groupId=${clientGroupId}`);

  ws.onmessage = function (e) {
    const messages = e.data.split('\n');
    for (let i = 0; i < messages.length; i++) {
      toast(messages[i]);
    }
  };

  ws.onclose = function () {
    setTimeout(function () {
      connectNoticeMe(clientId, clientGroupId);
    }, 3000);
  };

  ws.onerror = function () {
    ws.close();
  };

  return true;
}