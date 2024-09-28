import {Button, TextField} from "@mui/material";
import {FormEvent} from "react";
import {Grid} from "@mui/system";
import Textarea from '@mui/joy/Textarea';
import {api} from "../util/api";
import {Typography} from "@mui/joy";

type PublishNotificationFormProps = {
  clientId: string;
}

export default function PublishNotificationForm({clientId}: PublishNotificationFormProps) {
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
        <Grid mb={2} container spacing={2}>
          <Typography variant="subtitle1" component="p">Your User ID is <b>{clientId}</b>. Change Client Id
            input value to "*" to publish to all clients on this server.</Typography>
        </Grid>

        <Grid mb={2} container spacing={2}>
          <TextField
            name="clientId"
            label="Client ID"
            variant="outlined"
            defaultValue={clientId}
            required
          />
        </Grid>
        <Grid mb={2} container spacing={2}>
          <TextField
            name="clientGroupId"
            label="Client Group ID"
            variant="outlined"
            defaultValue={"*"}
            required
          />
        </Grid>
        <Grid mb={2} container spacing={2}>
          <Textarea
            size={"lg"}
            placeholder={"Write your notification body"}
            name="body"
            required
            defaultValue={'foo bar'}
          />
        </Grid>

        <Grid mb={2} container spacing={2}>
          <Button type="submit" variant="contained" color="primary">
            Publish Notification
          </Button>
        </Grid>
      </form>
    </>
  );
}
