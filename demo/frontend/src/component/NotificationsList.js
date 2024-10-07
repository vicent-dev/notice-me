import {useEffect, useState} from "react";
import {DataGrid} from '@mui/x-data-grid';
import {api} from "../util/api";
import {Button} from "@mui/material";
import toast from "react-hot-toast";
import {CircularProgress} from "@mui/joy";

export default function NotificationsList({refreshNotifications, setRefreshNotifications}) {
  const [notifications, setNotifications] = useState(null);

  function fetchNotifications() {
    //@todo change to have backend pagination and not only last 50
    api().get('/notifications?page=1&pageSize=50')
      .then((result) => {
        if (result && result.data) {
          setNotifications(result.data.rows);
        } else {
          setNotifications([]);
        }
      })
      .catch((error) => {
        console.log(error)
        toast.error('Something went wrong, try again later');
        setNotifications([]);
      });
  }

  useEffect(() => {
    if (refreshNotifications) {
        setRefreshNotifications(false);
        setNotifications(null);
        fetchNotifications();
    }
  }, [notifications, setRefreshNotifications, refreshNotifications]);

  function deleteNotification(id){
    api()
      .delete(`/notifications/${id}`)
      .then(() => {
        setRefreshNotifications(true);
      })
      .catch((error) => {
        console.log(error)
        toast.error('Something went wrong, try again later');
      });
  }


  const columns = [
    {
      field: 'ClientId',
      headerName: 'Client Id',
      width: 150,
      sortable: true
    },
    {
      field: 'ClientGroupId',
      headerName: 'Client Group Id',
      width: 150,
      sortable: true
    },
    {
      field: 'Body',
      headerName: 'Body',
      width: 150,
    },
    {
      field: 'CreatedAt',
      headerName: 'Created At',
      width: 250,
    },
    {
      field: 'NotifiedAt',
      headerName: 'Notified At',
      width: 250,
    },
    {
      field: 'ID',
      headerName: '',
      renderCell: (params) => (
        <Button
          variant="contained"
          size="small"
          color={'warning'}
          onClick={() => deleteNotification(params.value)}
        >
          Delete
        </Button>
      )
    }
  ];

  return (
    <>
      <h2>Notifications list </h2>
      {
        null === notifications ? (
          <CircularProgress/>
        ) : (
          <div style={{ height: 400, width: '100%' }}>
            <DataGrid
              autoPageSize={true}
              rows={notifications}
              columns={columns}
              initialState={{
                sorting: {
                  sortModel: [{ field: 'CreatedAt', sort: 'desc' }],
                },
              }}
              checkboxSelection
              disableRowSelectionOnClick
              getRowId={(row: any) => row.ID}
            />
          </div>
        )
      }
    </>
  );
}