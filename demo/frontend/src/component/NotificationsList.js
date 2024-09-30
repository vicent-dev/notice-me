import {useEffect, useState} from "react";
import {DataGrid, GridColDef} from '@mui/x-data-grid';
import {api} from "../util/api";
import {Button} from "@mui/material";
import toast from "react-hot-toast";
import {Box, CircularProgress} from "@mui/joy";

export default function NotificationsList({refreshNotifications, setRefreshNotifications}) {
  const [notifications, setNotifications] = useState(null);

  function fetchNotifications() {
    api().get('/notifications')
      .then((result) => {
        setNotifications(result.data);
      })
      .catch((error) => {
        toast.error(error)
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
      .catch((error) => toast.error(error));
  }


  const columns: GridColDef<>[] = [
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
      renderCell: (params) =>  (
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

  if (notifications === null) {
    return <CircularProgress />
  }

  return ( notifications.length > 0 && (
    <Box mt={2}>
      <h2>Notifications list </h2>
      <DataGrid
        rows={notifications}
        columns={columns}
        initialState={{
          pagination: {
            paginationModel: {
              pageSize: 100,
            },
          },
        }}
        pageSizeOptions={[5]}
        checkboxSelection
        disableRowSelectionOnClick
        getRowId={(row: any) => row.ID}
      />
    </Box>
  ));
}