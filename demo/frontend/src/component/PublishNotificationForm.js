import {Button, TextField} from "@mui/material";
import {Grid} from "@mui/system";
import {api} from "../util/api";
import '../index.css';
import toast from "react-hot-toast";

export default function PublishNotificationForm({clientId, clientGroupId, setRefreshNotifications}) {
  const handleSubmit = (event) => {
    event.preventDefault();

    api().post('/notifications', {
      "clientId": event.target.clientId.value,
      "clientGroupId": event.target.clientGroupId.value ?? '',
      "body": event.target.body.value,
    })
      .then(() => setRefreshNotifications(true))
      .catch((error) => {
        console.log(error);
        toast.error('Something went wrong, try again later');
      });
  }

  return (
    <>
      <form onSubmit={handleSubmit}>
        <Grid mb={4} item spacing={2}>
          <h4>Change Client Id/Client Group Id input value to "*" to publish to all clients/groups on this server.</h4>
          <ul>
            <li>User ID: <b className={'pointerOnHover'}
                            onClick={() => navigator.clipboard.writeText(clientId)}>{clientId}</b></li>
            <li>Group ID: <b className={'pointerOnHover'}
                             onClick={() => navigator.clipboard.writeText(clientGroupId)}>{clientGroupId}</b></li>
          </ul>
        </Grid>

        <Grid mb={2} item>
          <TextField
            name="clientId"
            label="Client ID"
            variant="outlined"
            defaultValue={clientId}
            required
          />
        </Grid>
        <Grid mb={2} item>
          <TextField
            name="clientGroupId"
            label="Client Group ID"
            variant="outlined"
            defaultValue={clientGroupId}
          />
        </Grid>
        <Grid mb={2} item>
          <TextField
            name="body"
            label="Body"
            variant="outlined"
            placeholder={"Write your notification body"}
            defaultValue={'foo bar'}
            required
          />
        </Grid>

        <Grid mb={2} item spacing={2}>
          <Button type="submit" variant="contained" color="primary">
            Publish Notification
          </Button>
        </Grid>
      </form>
    </>
  );
}
