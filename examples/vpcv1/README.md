Note that the ___classicAccess___ field is <ins>compulsory</ins>.

It is actually optional in the cloud API spec... (and assumes a default value of _false_ upon creation). 

However, if we had kept it optional, we would be allowed to have a yaml file with an empty _forProvider_ section (as all other fields - including _name_ - are optional); crossplane does not like that (we briefly considered creating an extra, non-optional "dummy" parameter, to get around the crossplane validator requirement, but decided to use the default-value-of-an-existing-parameter instead).
