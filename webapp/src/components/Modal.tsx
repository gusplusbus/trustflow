import * as React from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  Divider,
  Box,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";

export interface ModalProps {
  open: boolean;
  title?: string;
  children?: React.ReactNode;
  actions?: React.ReactNode; // Footer buttons (optional)
  maxWidth?: "xs" | "sm" | "md" | "lg" | "xl";
  fullWidth?: boolean;
  onClose: () => void;
}

export default function Modal({
  open,
  title,
  children,
  actions,
  maxWidth = "sm",
  fullWidth = true,
  onClose,
}: ModalProps) {
  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth={maxWidth}
      fullWidth={fullWidth}
    >
      {/* Header */}
      {title && (
        <>
          <DialogTitle sx={{ m: 0, p: 2 }}>
            {title}
            <IconButton
              aria-label="close"
              onClick={onClose}
              sx={{
                position: "absolute",
                right: 8,
                top: 8,
                color: (theme) => theme.palette.grey[500],
              }}
            >
              <CloseIcon />
            </IconButton>
          </DialogTitle>
          <Divider />
        </>
      )}

      {/* Body */}
      <DialogContent dividers>
        <Box sx={{ py: 1 }}>{children}</Box>
      </DialogContent>

      {/* Footer */}
      {actions && (
        <>
          <Divider />
          <DialogActions>{actions}</DialogActions>
        </>
      )}
    </Dialog>
  );
}
