# Deleting an Ensembler with related active entities

This page describes the process of deleting ensemblers **with related active entities** (ensembling jobs and/or router versions).

Ensemblers with related active router versions or ensembling jobs cannot be deleted. Ensemblers that are currently used by any routers also cannot be deleted.

Navigate to the Ensemblers page. Click on the delete button:

![](../../.gitbook/assets/ensembler_page.png)

The ensembler cannot be deleted since there are active router versions or ensembling jobs using the ensembler. Hence, the dialog will show the related entity that blocks the deletion process:

![](../../.gitbook/assets/delete_ensembler_modal_active.png)

If you still wish to delete the ensembler, please follow the instructions shown in the dialog.