import {Button, TextField} from "@mui/material";
import {FormEvent} from "react";
import {Grid} from "@mui/system";
import Textarea from '@mui/joy/Textarea';
import {api} from "../util/api";
import {Typography} from "@mui/joy";

type PublishNotificationFormProps = {
  clientId: string;
  clientGroupId: string;
}

export default function PublishNotificationForm({clientId, clientGroupId}: PublishNotificationFormProps) {
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    api().post('/notification', {
      "clientId": event.target.clientId.value,
      "clientGroupId": event.target.clientGroupId.value,
      "body": event.target.body.value,
    })
      .then()
      .catch((error) => {
        console.log(error);
      });
  }

  return (
    <>
      <form onSubmit={handleSubmit}>
        <Grid mb={2} row spacing={2}>
          <ul>
            <li>User ID: {clientId}</li>
            <li>Group ID: {clientGroupId}</li>
          </ul>
        </Grid>

        <Grid mb={2} row spacing={2}>
          <Typography component="p">Change Client Id/Client Group Id input value to "*" to publish to all clients/groups on this server.</Typography>
        </Grid>

        <Grid mb={2} row spacing={2}>
          <TextField
            name="clientId"
            label="Client ID"
            variant="outlined"
            defaultValue={clientId}
            required
          />
        </Grid>
        <Grid mb={2} row spacing={2}>
          <TextField
            name="clientGroupId"
            label="Client Group ID"
            variant="outlined"
            defaultValue={clientGroupId}
            required
          />
        </Grid>
        <Grid mb={2} row spacing={2}>
          <Textarea
            size={"lg"}
            placeholder={"Write your notification body"}
            name="body"
            required
            defaultValue={'foo bar'}
          />
        </Grid>

        <Grid mb={2} row spacing={2}>
          <Button type="submit" variant="contained" color="primary">
            Publish Notification
          </Button>
        </Grid>
      </form>
    </>
  );
}
