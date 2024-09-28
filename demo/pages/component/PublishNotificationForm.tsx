import {Button, TextField, Typography} from "@mui/material";
import {api} from "@/pages/util/api";
import {FormEvent} from "react";
import {Grid} from "@mui/system";
import Textarea from '@mui/joy/Textarea';
import connectNoticeMe from "@/pages/util/websocket";

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
        <Typography variant="subtitle1" component="p">Your User ID is {clientId}. Change Client Id input value to "*" to publish to all clients on this server.</Typography>
      </Grid>

      <Grid mb={2} container spacing={2}>
        <TextField
          name="clientId"
          label="Client ID"
          variant="outlined"
          defaultValue={clientId}
          required
        />
        <TextField
          name="clientGroupId"
          label="Client Group ID"
          variant="outlined"
          defaultValue={"*"}
          required
        />
        <Textarea
          size={"lg"}
          placeholder={"Write your notification body"}
          name="body"
          required
        />
      </Grid>

      <Button type="submit" variant="contained" color="primary">
        Publish Notification
      </Button>
    </form>
    </>
  );
}
