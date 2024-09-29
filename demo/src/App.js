import './App.css';
import {Grid} from "@mui/system";
import {Box} from "@mui/joy";
import NoticeMe from "./component/NoticeMe";
import PublishNotificationForm from "./component/PublishNotificationForm";

function App() {
  const clientId = crypto.randomUUID();

  const groupIds = [
    'group_1',
    'group_2',
    'group_3',
    'group_4',
    'group_5',
    'group_6',
    'group_7',
    'group_8',
  ];

  var clientGroupId = groupIds[Math.floor(Math.random() * groupIds.length)];

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
        <PublishNotificationForm clientId={clientId} clientGroupId={clientGroupId}/>
      </Grid>
      <NoticeMe clientId={clientId} clientGroupId={clientGroupId}/>
    </>
  );
}

export default App;
