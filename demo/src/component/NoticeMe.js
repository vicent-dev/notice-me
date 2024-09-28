import {getWebsocketBaseURL} from "../util/api";
import toast, { Toaster } from 'react-hot-toast';
import React, {useEffect, useState} from "react";

type NoticeMeProps = {
  clientId: string;
  clientGroupId: null|string;
};

export default function NoticeMe({clientId, clientGroupId}: NoticeMeProps) {

  const [ws, setWs] = useState(null);

  useEffect(() => {
    const wsCnn = new WebSocket(`${getWebsocketBaseURL()}/ws?id=${clientId}&groupId=${clientGroupId}`);

    wsCnn.onopen = () => {
      setWs(wsCnn);
    };

    return () => {
      wsCnn.close();
    }
  }, []);

  useEffect(() => {
    if (!ws) return;

    ws.onmessage = e => {
      toast.success((t) => <>
        <span dangerouslySetInnerHTML={{__html: e.data}}></span>
      </>);
    };
  }, [ws]);


  return <>
    <Toaster
        position="top-right"
        reverseOrder={false}
    />
  </>;
}