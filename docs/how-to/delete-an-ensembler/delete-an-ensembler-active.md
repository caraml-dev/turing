# Deleting an Ensembler with related active entity

An Ensembler only could be deleted if there are no active ensembling jobs / router version that use the ensembler. This page show the deletion of ensembler **with related active entity** (router version and ensembling job).

Ensembler with related active router version or ensembling jobs can not be deleted. Ensembler that currently used by router is also can not be deleted.

Navigate to Ensemblers page

Click on the delete button 

![](../../.gitbook/assets/ensembler_page.png)

The ensembler can not be deleted since there are active router version/ensembling job using the ensembler. Hence, the modal will show the related entity that block the deletion process.

![](../../.gitbook/assets/delete_ensembler_modal_active.png)

If you still wish to delete the ensembler, please follow notes on the modal. Since there are multiple constraint for the ensembler deletion process.