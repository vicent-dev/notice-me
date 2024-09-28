import {Box, Typography} from "@mui/material";
import {Grid} from "@mui/system";
import PublishNotificationForm from "@/pages/component/PublishNotificationForm";
import connectNoticeMe from "@/pages/util/websocket";
import {useEffect, useState} from "react";

export default function Home() {
  const clientId = crypto.randomUUID();
  const [connectedWs, setConnectedWs] = useState<boolean>(false);

  useEffect(() => {
    if(!connectedWs) {
      setConnectedWs(connectNoticeMe(clientId));
    }
  }, [])

  return (
    <>
      <Grid
        container
        spacing={0}
        direction="column"
        alignItems="center"
        justifyContent="center"
      >
        <Box mb={2}>
          <Typography variant="h3" component="h1">Noticeme</Typography>
        </Box>
        <PublishNotificationForm clientId={clientId}/>
      </Grid>
    </>
  );
}
