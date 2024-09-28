import './App.css';
import {Grid} from "@mui/system";
import {Box, Typography} from "@mui/joy";
import NoticeMe from "./component/NoticeMe";
import PublishNotificationForm from "./component/PublishNotificationForm";

function App() {
  const clientId = crypto.randomUUID();

  return (
      <>
        <Grid
            container
            direction="column"
            alignItems="center"
            justifyContent="center"
        >
          <Box mb={2}>
            <h1>Noticeme</h1>
          </Box>
          <PublishNotificationForm clientId={clientId}/>
        </Grid>
        <NoticeMe clientId={clientId} clientGroupId={'*'}/>
      </>
  );
}

export default App;
