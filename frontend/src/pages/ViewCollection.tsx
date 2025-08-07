import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import {
  Box,
  Typography,
  CircularProgress,
  Alert,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  TablePagination,
} from '@mui/material';
import itemsService, { Item } from '../services/itemsService';
import collectionsService, { CollectionWithFields } from '../services/collectionsService';

const ViewCollection: React.FC = () => {
  const { collectionName } = useParams<{ collectionName: string }>();
  const [collection, setCollection] = useState<CollectionWithFields | null>(null);
  const [items, setItems] = useState<Item[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [totalItems, setTotalItems] = useState(0);

  useEffect(() => {
    if (collectionName) {
      fetchData();
    }
  }, [collectionName, page, rowsPerPage]);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      // Fetch collection details
      const collectionDetails = await collectionsService.getCollection(collectionName!);
      setCollection(collectionDetails);

      // Fetch items
      const itemsResponse = await itemsService.getItems(collectionName!, page + 1, rowsPerPage);
      setItems(itemsResponse.data);
      setTotalItems(itemsResponse.meta.total);

    } catch (err) {
      setError('Failed to load collection data.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  if (!collection) {
    return <Typography>Collection not found.</Typography>;
  }

  const { fields } = collection;
  const visibleFields = fields.filter(field => !field.hidden);

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Items in {collection.collection.collection}
      </Typography>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              {visibleFields.map((field) => (
                <TableCell key={field.field}>{field.field}</TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id}>
                {visibleFields.map((field) => (
                  <TableCell key={field.field}>{item[field.field]}</TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        rowsPerPageOptions={[5, 10, 25]}
        component="div"
        count={totalItems}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </Box>
  );
};

export default ViewCollection;
