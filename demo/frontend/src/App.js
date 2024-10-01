import './App.css';
import {Grid} from "@mui/system";
import {Box, Divider} from "@mui/joy";
import NoticeMe from "./component/NoticeMe";
import PublishNotificationForm from "./component/PublishNotificationForm";
import NotificationsList from './component/NotificationsList';
import {useState} from "react";

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

const clientGroupId = groupIds[Math.floor(Math.random() * groupIds.length)];

function App() {
  const [refreshNotifications, setRefreshNotifications] = useState(true);

  return (
    <>
      <Grid
        container
        direction="column"
        justifyContent="center"
      >
        <Box mb={2}>
          <h1>Noticeme</h1>
        </Box>
        <PublishNotificationForm
          clientId={clientId}
          clientGroupId={clientGroupId}
          setRefreshNotifications={setRefreshNotifications}
        />
        <Divider />
        <NotificationsList
          refreshNotifications={refreshNotifications}
          setRefreshNotifications={setRefreshNotifications}
        />
      </Grid>
      <NoticeMe
        clientId={clientId}
        clientGroupId={clientGroupId}
        setRefreshNotifications={setRefreshNotifications}
      />
    </>
  );
}

export default App;
